package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

const (
	scenariosDir = "../scenarios"
	fixturesDir  = "../fixtures"
)

func TestExpandCommand(t *testing.T) {
	dir := t.TempDir()
	cmd := "---\ndescription: Match the task to a playbook.\nagent: software-engineer\n---\n\nLoad the `drill` skill, then run it for this task:\n\n$ARGUMENTS\n"
	if err := os.WriteFile(filepath.Join(dir, "drill.md"), []byte(cmd), 0o644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name, prompt, want string
	}{
		{"expands arguments", "/drill fix the bug", "Load the `drill` skill, then run it for this task:\n\nfix the bug"},
		{"keeps multi-line arguments", "/drill line one\nline two", "Load the `drill` skill, then run it for this task:\n\nline one\nline two"},
		{"leaves plain prompts alone", "how does Remove work?", "how does Remove work?"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := expandCommand(tt.prompt, dir)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("expandCommand(%q) =\n%q\nwant\n%q", tt.prompt, got, tt.want)
			}
		})
	}

	if _, err := expandCommand("/missing task", dir); err == nil {
		t.Fatal("expandCommand with an unknown command returned no error")
	}
}

func TestDefaultJobsStaysBetweenThreeAndSix(t *testing.T) {
	for runs, want := range map[int]int{1: 3, 3: 3, 5: 5, 6: 6, 24: 6} {
		if got := defaultJobs(runs); got != want {
			t.Errorf("defaultJobs(%d) = %d, want %d", runs, got, want)
		}
	}
}

func TestParseIDs(t *testing.T) {
	got, err := parseIDs("1, 4-6", scenariosDir)
	if err != nil {
		t.Fatal(err)
	}
	if want := []int{1, 4, 5, 6}; !slices.Equal(got, want) {
		t.Fatalf("parseIDs = %v, want %v", got, want)
	}
	if _, err := parseIDs("6-4", scenariosDir); err == nil {
		t.Fatal("accepted a descending range")
	}
}

// Every scenario must load, build from its fixture, survive its setup.sh, and
// leave Go code that compiles, so a broken fixture fails here instead of in a
// paid run.
func TestEveryScenarioBuilds(t *testing.T) {
	ids, err := parseIDs("all", scenariosDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) == 0 {
		t.Fatal("no scenarios found")
	}
	for _, id := range ids {
		sc, err := loadScenario(scenariosDir, fixturesDir, id)
		if err != nil {
			t.Fatal(err)
		}
		t.Run(sc.name, func(t *testing.T) {
			t.Parallel()
			dest := filepath.Join(t.TempDir(), "sandbox")
			if _, err := buildSandbox(sc, dest, "test", ""); err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(filepath.Join(dest, "go.mod")); err != nil {
				return
			}
			build := exec.Command("go", "build", "./...")
			build.Dir = dest
			if out, err := build.CombinedOutput(); err != nil {
				t.Fatalf("go build: %v\n%s", err, out)
			}
		})
	}
}

func TestRemoteSandboxGetsItsOwnBaseAndDraft(t *testing.T) {
	sc, err := loadScenario(scenariosDir, fixturesDir, 24)
	if err != nil {
		t.Fatal(err)
	}
	if !sc.remote || !sc.draftPR {
		t.Fatal("scenario 024 must start from an existing draft PR")
	}
	tmp := t.TempDir()
	remote, dest := filepath.Join(tmp, "origin.git"), filepath.Join(tmp, "sandbox")
	if _, err := git(tmp, "init", "--quiet", "--bare", remote); err != nil {
		t.Fatal(err)
	}

	base, err := buildSandbox(sc, dest, "r1-024-1", remote)
	if err != nil {
		t.Fatal(err)
	}
	if base != "base-r1-024-1" {
		t.Fatalf("base = %q, want base-r1-024-1", base)
	}

	// Record the PR request instead of calling GitHub.
	if err := os.WriteFile(filepath.Join(tmp, "gh"), []byte("#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$PR_ARGS\"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", tmp+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("PR_ARGS", filepath.Join(tmp, "args"))
	if err := openDraftPR(context.Background(), dest, base); err != nil {
		t.Fatal(err)
	}

	head := "remove-lookup-r1-024-1"
	out, err := git(remote, "rev-list", "--count", base+".."+head)
	if err != nil || strings.TrimSpace(out) != "1" {
		t.Fatalf("remote head is %q commits ahead of base: %v", out, err)
	}
	args, err := os.ReadFile(filepath.Join(tmp, "args"))
	want := "pr\ncreate\n--draft\n--fill\n--base\n" + base + "\n--head\n" + head + "\n"
	if err != nil || string(args) != want {
		t.Fatalf("draft request = %q (%v), want %q", args, err, want)
	}
}

func TestDraftPRNeedsRemote(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "001-x")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{
		"prompt.md":     "task",
		"scenario.json": `{"fixture": "tally", "draftPR": true}`,
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := loadScenario(root, fixturesDir, 1); err == nil {
		t.Fatal("accepted draftPR without remote")
	}
}
