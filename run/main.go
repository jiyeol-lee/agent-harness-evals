// Command run builds a throwaway Git sandbox for each scenario, sends the
// scenario's prompts to `opencode run`, and writes transcripts and a summary
// for grading. Run it from the repository root:
//
//	go run ./run [flags]
//
// Each scenario lives in scenarios/<NNN>-<slug>/:
//
//	scenario.json  {"fixture": "tally", "command": "drill", "remote": false, "draftPR": false}
//	prompt.md      first message; prompt-2.md, prompt-3.md, ... follow in the same session
//	setup.sh       optional; runs in the sandbox after the fixture's initial commit
//	checks.md      what the grader verifies; the runner does not read it
//
// A sandbox is a copy of fixtures/<fixture>/ committed as "Initial commit" on
// main. Sandboxes go to /tmp/agentic-coding-tool/<fixture>-<tag>, where tag is
// <run>-<NNN>-<repeat>; the path avoids the word "eval" so the agent under test
// is not told it is being tested. Transcripts and summary.md go to
// /tmp/agentic-coding-tool/evals-<run>/.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const tmpRoot = "/tmp/agentic-coding-tool"

type options struct {
	scenarios string
	fixtures  string
	ids       []int
	repeats   int
	jobs      int
	agent     string
	model     string
	remote    string
	commands  string
	runID     string
	timeout   time.Duration
	buildOnly bool
}

type scenario struct {
	id      int
	name    string
	dir     string
	fixture string
	prompts []string
	remote  bool
	draftPR bool
}

type job struct {
	sc     scenario
	repeat int
}

type result struct {
	sc       scenario
	repeat   int
	sandbox  string
	session  string
	duration time.Duration
	skills   []string
	subagent int
	status   string
	note     string
}

func main() {
	opts, err := parseFlags()
	if err != nil {
		fail(err)
	}

	var local, remote []job
	var skipped []result
	for _, id := range opts.ids {
		sc, err := loadScenario(opts.scenarios, opts.fixtures, id)
		if err != nil {
			fail(err)
		}
		for i, p := range sc.prompts {
			if sc.prompts[i], err = expandCommand(p, opts.commands); err != nil {
				fail(fmt.Errorf("scenario %03d: %w", id, err))
			}
		}
		for k := 1; k <= opts.repeats; k++ {
			j := job{sc: sc, repeat: k}
			switch {
			case !sc.remote:
				local = append(local, j)
			case opts.remote != "" || opts.buildOnly:
				remote = append(remote, j)
			default:
				skipped = append(skipped, result{sc: sc, repeat: k, note: "skipped: needs -remote"})
			}
		}
	}
	if opts.jobs == 0 {
		opts.jobs = defaultJobs(len(local) + len(remote))
	}

	logDir := filepath.Join(tmpRoot, "evals-"+opts.runID)
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		fail(err)
	}
	fmt.Printf("run %s: %d run(s), %d in parallel, %d skipped, logs in %s\n",
		opts.runID, len(local)+len(remote), opts.jobs, len(skipped), logDir)

	all := append(skipped, runAll(opts, logDir, local, remote)...)
	sort.Slice(all, func(i, j int) bool {
		if all[i].sc.id != all[j].sc.id {
			return all[i].sc.id < all[j].sc.id
		}
		return all[i].repeat < all[j].repeat
	})
	summary := filepath.Join(logDir, "summary.md")
	if err := os.WriteFile(summary, []byte(renderSummary(opts, all)), 0o644); err != nil {
		fail(err)
	}
	fmt.Println("summary:", summary)
}

// runAll runs local jobs on opts.jobs workers. Remote jobs share one GitHub
// repository, where parallel agents could push over each other's branches, so
// the first worker runs them one at a time before it helps with local jobs.
func runAll(opts options, logDir string, local, remote []job) []result {
	localCh, remoteCh := make(chan job), make(chan job)
	results := make(chan result)
	var wg sync.WaitGroup
	for w := range opts.jobs {
		wg.Go(func() {
			if w == 0 {
				for j := range remoteCh {
					results <- runOne(opts, logDir, j)
				}
			}
			for j := range localCh {
				results <- runOne(opts, logDir, j)
			}
		})
	}
	go func() {
		for _, j := range remote {
			remoteCh <- j
		}
		close(remoteCh)
	}()
	go func() {
		for _, j := range local {
			localCh <- j
		}
		close(localCh)
	}()
	go func() {
		wg.Wait()
		close(results)
	}()

	var all []result
	for r := range results {
		state := "done"
		if r.note != "" {
			state = r.note
		}
		fmt.Printf("%s #%d %s in %s (%s)\n", r.sc.name, r.repeat, state, r.duration.Round(time.Second), r.sandbox)
		all = append(all, r)
	}
	return all
}

func parseFlags() (options, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return options{}, err
	}
	var opts options
	var ids string
	flag.StringVar(&ids, "s", "all", `scenarios to run, such as "all", "7", "1,5,19", or "3-8"`)
	flag.IntVar(&opts.repeats, "n", 1, "runs per scenario")
	flag.IntVar(&opts.jobs, "j", 0, "runs in parallel; 0 means the number of runs, clamped to 3-6")
	flag.StringVar(&opts.agent, "agent", "software-engineer", "agent passed to opencode run")
	flag.StringVar(&opts.model, "model", "", "model passed to opencode run, as provider/model; empty uses the agent's default")
	flag.StringVar(&opts.remote, "remote", "", "disposable GitHub repository URL; scenarios with \"remote\": true are skipped without it")
	flag.StringVar(&opts.commands, "commands", filepath.Join(home, ".config/opencode/command"), "directory of OpenCode commands used to expand a prompt that starts with /<name>")
	flag.StringVar(&opts.runID, "id", time.Now().Format("20060102-150405"), "run ID used in sandbox, branch, and log names")
	flag.DurationVar(&opts.timeout, "timeout", 90*time.Minute, "limit for each opencode run call")
	flag.BoolVar(&opts.buildOnly, "build-only", false, "build the sandboxes without running opencode or touching GitHub")
	flag.Parse()

	if opts.repeats < 1 || opts.jobs < 0 {
		return opts, errors.New("-n must be at least 1 and -j at least 0")
	}
	if opts.scenarios, err = filepath.Abs("scenarios"); err != nil {
		return opts, err
	}
	if opts.fixtures, err = filepath.Abs("fixtures"); err != nil {
		return opts, err
	}
	opts.ids, err = parseIDs(ids, opts.scenarios)
	return opts, err
}

func defaultJobs(runs int) int {
	return min(max(runs, 3), 6)
}

func parseIDs(spec, root string) ([]int, error) {
	if spec == "all" {
		dirs, err := scenarioDirs(root)
		if err != nil {
			return nil, err
		}
		ids := make([]int, 0, len(dirs))
		for id := range dirs {
			ids = append(ids, id)
		}
		sort.Ints(ids)
		return ids, nil
	}

	var ids []int
	for part := range strings.SplitSeq(spec, ",") {
		lo, hi, isRange := strings.Cut(strings.TrimSpace(part), "-")
		a, err := strconv.Atoi(lo)
		if err != nil {
			return nil, fmt.Errorf("bad scenario %q", part)
		}
		b := a
		if isRange {
			if b, err = strconv.Atoi(hi); err != nil || b < a {
				return nil, fmt.Errorf("bad scenario range %q", part)
			}
		}
		for n := a; n <= b; n++ {
			ids = append(ids, n)
		}
	}
	return ids, nil
}

// scenarioDirs maps each scenario ID to its directory name, such as
// 7 -> "007-investigation-order-trace".
func scenarioDirs(root string) (map[int]string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	dirs := map[int]string{}
	for _, e := range entries {
		prefix, _, _ := strings.Cut(e.Name(), "-")
		id, err := strconv.Atoi(prefix)
		if !e.IsDir() || len(prefix) != 3 || err != nil {
			continue
		}
		if other, dup := dirs[id]; dup {
			return nil, fmt.Errorf("scenarios %s and %s share ID %03d", other, e.Name(), id)
		}
		dirs[id] = e.Name()
	}
	return dirs, nil
}

func loadScenario(root, fixtures string, id int) (scenario, error) {
	// setup.sh runs from inside the sandbox, so relative paths would break.
	root, err := filepath.Abs(root)
	if err != nil {
		return scenario{}, err
	}
	if fixtures, err = filepath.Abs(fixtures); err != nil {
		return scenario{}, err
	}
	dirs, err := scenarioDirs(root)
	if err != nil {
		return scenario{}, err
	}
	name, ok := dirs[id]
	if !ok {
		return scenario{}, fmt.Errorf("no scenario %03d in %s", id, root)
	}
	sc := scenario{id: id, name: name, dir: filepath.Join(root, name)}

	raw, err := os.ReadFile(filepath.Join(sc.dir, "scenario.json"))
	if err != nil {
		return sc, err
	}
	var cfg struct {
		Fixture string `json:"fixture"`
		Command string `json:"command"`
		Remote  bool   `json:"remote"`
		DraftPR bool   `json:"draftPR"`
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&cfg); err != nil {
		return sc, fmt.Errorf("%s/scenario.json: %w", name, err)
	}
	if cfg.Fixture == "" || strings.ContainsAny(cfg.Fixture, `/\`) {
		return sc, fmt.Errorf("%s/scenario.json: fixture must name a directory in %s", name, fixtures)
	}
	sc.fixture = filepath.Join(fixtures, cfg.Fixture)
	if info, err := os.Stat(sc.fixture); err != nil || !info.IsDir() {
		return sc, fmt.Errorf("%s: fixture %q not found", name, cfg.Fixture)
	}
	if strings.ContainsAny(cfg.Command, "/\\ \t\r\n") {
		return sc, fmt.Errorf("%s: invalid command name %q", name, cfg.Command)
	}
	if cfg.DraftPR && !cfg.Remote {
		return sc, fmt.Errorf("%s: draftPR needs remote: true", name)
	}
	sc.remote, sc.draftPR = cfg.Remote, cfg.DraftPR

	for i := 1; ; i++ {
		file := "prompt.md"
		if i > 1 {
			file = fmt.Sprintf("prompt-%d.md", i)
		}
		data, err := os.ReadFile(filepath.Join(sc.dir, file))
		if errors.Is(err, fs.ErrNotExist) && i > 1 {
			break
		}
		if err != nil {
			return sc, err
		}
		sc.prompts = append(sc.prompts, strings.TrimSpace(string(data)))
	}
	if cfg.Command != "" {
		sc.prompts[0] = fmt.Sprintf("/%s %s", cfg.Command, sc.prompts[0])
	}
	return sc, nil
}

func runOne(opts options, logDir string, j job) (r result) {
	start := time.Now()
	sc := j.sc
	tag := fmt.Sprintf("%s-%03d-%d", opts.runID, sc.id, j.repeat)
	r = result{sc: sc, repeat: j.repeat}
	r.sandbox = filepath.Join(tmpRoot, fmt.Sprintf("%s-%s", filepath.Base(sc.fixture), tag))
	defer func() { r.duration = time.Since(start) }()

	remote := opts.remote
	if opts.buildOnly || !sc.remote {
		remote = ""
	}
	base, err := buildSandbox(sc, r.sandbox, tag, remote)
	if err != nil {
		r.note = "build sandbox: " + err.Error()
		return r
	}
	if !opts.buildOnly {
		r.note = runAgent(opts, logDir, sc, j.repeat, base, &r)
	}
	if out, err := git(r.sandbox, "status", "--short", "--branch"); err == nil {
		r.status = strings.TrimSpace(out)
	}
	return r
}

// runAgent sends every prompt in one OpenCode session and records what the
// transcript shows. It returns a note when the run did not finish cleanly.
func runAgent(opts options, logDir string, sc scenario, repeat int, base string, r *result) string {
	ctx, cancel := context.WithTimeout(context.Background(), opts.timeout)
	defer cancel()
	if sc.draftPR {
		if err := openDraftPR(ctx, r.sandbox, base); err != nil {
			return "open draft PR: " + err.Error()
		}
	}

	logBase := filepath.Join(logDir, fmt.Sprintf("%03d-%d", sc.id, repeat))
	stdout, err := os.Create(logBase + ".jsonl")
	if err != nil {
		return err.Error()
	}
	defer stdout.Close()
	stderr, err := os.Create(logBase + ".err")
	if err != nil {
		return err.Error()
	}
	defer stderr.Close()

	var note string
	for i, prompt := range sc.prompts {
		args := []string{"run", "--archived", "--auto", "--format", "json", "--agent", opts.agent}
		if opts.model != "" {
			args = append(args, "--model", opts.model)
		}
		if i > 0 {
			if r.session == "" {
				note = "no session ID from the first prompt, so follow-ups were not sent"
				break
			}
			args = append(args, "--session", r.session)
		}
		args = append(args, strings.ReplaceAll(prompt, "{{base}}", base))

		turnCtx, turnCancel := context.WithTimeout(context.Background(), opts.timeout)
		var turn bytes.Buffer
		cmd := exec.CommandContext(turnCtx, "opencode", args...)
		cmd.Dir = r.sandbox
		cmd.Stdout = io.MultiWriter(stdout, &turn)
		cmd.Stderr = stderr
		runErr := cmd.Run()
		turnCancel()

		if r.session == "" {
			r.session = firstSession(turn.Bytes())
		}
		if runErr != nil {
			note = fmt.Sprintf("prompt %d: %v", i+1, runErr)
			break
		}
	}

	transcript, _ := os.ReadFile(logBase + ".jsonl")
	r.skills = loadedSkills(transcript)
	r.subagent = bytes.Count(transcript, []byte(`"tool":"subagent"`))
	return note
}

// fixtureEnv pins the identity and dates of fixture commits so that every
// sandbox of a scenario starts from the same commit IDs, and keeps the user's
// signing setup out of them.
var fixtureEnv = []string{
	"GIT_AUTHOR_NAME=Fixture",
	"GIT_AUTHOR_EMAIL=fixture@example.invalid",
	"GIT_COMMITTER_NAME=Fixture",
	"GIT_COMMITTER_EMAIL=fixture@example.invalid",
	"GIT_AUTHOR_DATE=2026-01-01T12:00:00Z",
	"GIT_COMMITTER_DATE=2026-01-01T12:00:00Z",
	"GIT_CONFIG_COUNT=2",
	"GIT_CONFIG_KEY_0=commit.gpgsign",
	"GIT_CONFIG_VALUE_0=false",
	"GIT_CONFIG_KEY_1=tag.gpgsign",
	"GIT_CONFIG_VALUE_1=false",
}

// buildSandbox copies the fixture to dest, commits it on main, publishes main
// as the scenario's own base branch when remote is set, and then runs the
// scenario's setup.sh. It returns the base branch name, or "" without a remote.
func buildSandbox(sc scenario, dest, tag, remote string) (string, error) {
	if _, err := os.Stat(dest); err == nil {
		return "", fmt.Errorf("%s already exists", dest)
	}
	if err := copyTree(sc.fixture, dest); err != nil {
		return "", err
	}
	for _, args := range [][]string{
		{"init", "--quiet", "--initial-branch=main"},
		{"add", "--all"},
		{"commit", "--quiet", "--message", "Initial commit"},
	} {
		if _, err := git(dest, args...); err != nil {
			return "", err
		}
	}

	base := ""
	if remote != "" {
		base = "base-" + tag
		for _, args := range [][]string{
			{"remote", "add", "origin", remote},
			{"push", "--quiet", "origin", "main:refs/heads/" + base},
			{"fetch", "--quiet", "origin", base},
			// The remote's real default branch is unrelated to this fixture,
			// so make "the default branch" mean the scenario's base.
			{"remote", "set-head", "origin", base},
			{"branch", "--quiet", "--set-upstream-to=origin/" + base, "main"},
		} {
			if _, err := git(dest, args...); err != nil {
				return "", err
			}
		}
	}

	setup := filepath.Join(sc.dir, "setup.sh")
	if _, err := os.Stat(setup); err == nil {
		cmd := exec.Command("bash", setup)
		cmd.Dir = dest
		cmd.Env = append(os.Environ(), fixtureEnv...)
		cmd.Env = append(cmd.Env, "TAG="+tag, "BASE="+base)
		if out, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("setup.sh: %w: %s", err, out)
		}
	}
	return base, nil
}

// openDraftPR publishes the branch that setup.sh left checked out and opens a
// draft PR for it, so the agent starts from an existing PR.
func openDraftPR(ctx context.Context, dir, base string) error {
	branch, err := git(dir, "branch", "--show-current")
	if err != nil {
		return err
	}
	branch = strings.TrimSpace(branch)
	if branch == "" || branch == "main" {
		return errors.New("setup.sh must leave a topic branch checked out")
	}
	if _, err := git(dir, "push", "--quiet", "--set-upstream", "origin", branch); err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, "gh", "pr", "create", "--draft", "--fill", "--base", base, "--head", branch)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("gh pr create: %w: %s", err, out)
	}
	return nil
}

func copyTree(src, dest string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dest, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode().Perm())
	})
}

func git(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), fixtureEnv...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, out)
	}
	return string(out), nil
}

// expandCommand turns "/name args" into the body of <dir>/name.md with
// $ARGUMENTS replaced, the way the OpenCode TUI expands a slash command.
// `opencode run` does not expand commands, so the runner does it.
func expandCommand(prompt, dir string) (string, error) {
	rest, ok := strings.CutPrefix(prompt, "/")
	if !ok {
		return prompt, nil
	}
	name, args := rest, ""
	if i := strings.IndexAny(rest, " \n"); i >= 0 {
		name, args = rest[:i], strings.TrimSpace(rest[i+1:])
	}
	data, err := os.ReadFile(filepath.Join(dir, name+".md"))
	if err != nil {
		return "", fmt.Errorf("command /%s: %w", name, err)
	}
	body := string(data)
	if after, ok := strings.CutPrefix(body, "---\n"); ok {
		if _, content, found := strings.Cut(after, "\n---\n"); found {
			body = content
		}
	}
	return strings.TrimSpace(strings.ReplaceAll(body, "$ARGUMENTS", args)), nil
}

var (
	sessionRE = regexp.MustCompile(`"sessionID":"(ses_[A-Za-z0-9]+)"`)
	skillRE   = regexp.MustCompile(`"tool":"skill","state":\{"status":"[a-z]+","input":\{"id":"([a-z0-9-]+)"\}`)
)

func firstSession(transcript []byte) string {
	if m := sessionRE.FindSubmatch(transcript); m != nil {
		return string(m[1])
	}
	return ""
}

func loadedSkills(transcript []byte) []string {
	var skills []string
	seen := map[string]bool{}
	for _, m := range skillRE.FindAllSubmatch(transcript, -1) {
		if s := string(m[1]); !seen[s] {
			seen[s] = true
			skills = append(skills, s)
		}
	}
	return skills
}

func renderSummary(opts options, results []result) string {
	var b strings.Builder
	model := opts.model
	if model == "" {
		model = "agent default"
	}
	fmt.Fprintf(&b, "# Run %s\n\n", opts.runID)
	fmt.Fprintf(&b, "Agent `%s`, model `%s`, %d run(s) per scenario. Grade each run against `%s/<scenario>/checks.md`.\n\n", opts.agent, model, opts.repeats, opts.scenarios)
	b.WriteString("| Scenario | Run | Time | Session | Skills loaded | Subagents | Git status | Sandbox | Note |\n")
	b.WriteString("| --- | --- | --- | --- | --- | --- | --- | --- | --- |\n")
	for _, r := range results {
		fmt.Fprintf(&b, "| %s | %d | %s | `%s` | %s | %d | %s | `%s` | %s |\n",
			r.sc.name, r.repeat, r.duration.Round(time.Second), r.session,
			cell(strings.Join(r.skills, ", ")), r.subagent, cell(r.status), r.sandbox, cell(r.note))
	}
	return b.String()
}

func cell(s string) string {
	s = strings.ReplaceAll(s, "|", `\|`)
	return strings.ReplaceAll(s, "\n", "<br>")
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "run:", err)
	os.Exit(1)
}
