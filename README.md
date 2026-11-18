# agent-harness-evals

Checks that my OpenCode harness (the `drill` command, its playbooks, and the skills they call) does what it promises. Each scenario sends a request to `opencode run` inside a throwaway Git sandbox. You grade the transcript and sandbox against the scenario's `checks.md`.

## How to run

Run these commands from the repository root:

```sh
go run ./run                                # all local scenarios
go run ./run -s 1,4-6 -n 3                  # scenarios 1 and 4 to 6, three runs each
go run ./run -model openai/gpt-6.2-sol      # pick the model
go run ./run -build-only                    # build the sandboxes only
go run ./run -remote <disposable-repo-url>  # also run the GitHub scenarios
```

Runs go in parallel: `-j` defaults to the number of runs, clamped to 3 to 6. Live runs need authenticated OpenCode. See `go run ./run -h` for all flags.

**GitHub scenarios make real changes.** Scenarios with `"remote": true` are skipped unless you pass `-remote` with a disposable repository. Authenticate `git` and `gh`, and enable Actions for scenario 024. These scenarios run one at a time, each against its own base branch named `base-<run>-<NNN>-<repeat>`. Afterwards, close the PRs and issues and delete the branches the run created. Leftover branches can collide with the next run.

## How to grade

1. Open `/tmp/agentic-coding-tool/evals-<run>/summary.md`. It lists each run's transcript, session, and sandbox.
2. Check routing, action order, and permissions in the transcript. Verify results in the sandbox or on the actual PR head. Don't grade reply headings or skill counts. Flag a principle load only if its trigger didn't apply.
3. Record **pass** if the checks hold, **fail** if one was missed, or **inconclusive** if the environment prevented verification. Cite the evidence.

## Folder structure

```text
run/                  Go runner and its tests
fixtures/
  tally/              Small CLI, one stock package: simple and GitHub scenarios
  shop/               Medium CLI with stock, orders, pricing, and events: design and perf scenarios
  dotfiles/           OpenCode dotfiles: skill authoring
scenarios/NNN-slug/
  scenario.json       {"fixture": "tally", "command": "drill"}, plus optional "remote" and "draftPR"
  prompt.md           First message; prompt-2.md, ... are follow-ups in the same session
  setup.sh            Optional; runs in the sandbox after the fixture is committed on main
  checks.md           What to verify and how
```

A sandbox is a copy of the fixture, committed on `main`, at `/tmp/agentic-coding-tool/<fixture>-<run>-<NNN>-<repeat>/`. Every fixture is a one-shot CLI with no server or port, so sandboxes never conflict. In prompts, `{{base}}` becomes the scenario's base branch. With `"draftPR": true`, the runner pushes the branch that `setup.sh` left checked out and opens a draft PR for it before the first prompt.

## How to add, change, or remove a scenario

- **Add:** Pick an unused three-digit number and the smallest fixture that can show the behavior. Put branches, commits, or uncommitted edits in `setup.sh`, not in a new fixture.
- **Change:** Keep the prompt, `setup.sh`, and checks in sync. Change a fixture only if every scenario that uses it still holds.
- **Remove:** Delete the directory. Don't renumber the others.

Validate with `go test ./run`. It builds every scenario's sandbox and compiles its code.
