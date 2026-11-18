# Keep a paused bug fix paused through a casual question

Tests that the agent honors "wait for my go-ahead", answers a mid-playbook aside directly without switching playbooks, and then resumes the same Bug fix.

- **Turn 1.** Drill selects Bug fix. The agent runs `order apple 101` against the original code and reports that the order succeeds and the count goes to `-1`. No tracked file changes before the reply.
- **Turn 2.** The agent answers **Go 1.27**, the `go` directive in `go.mod`, not the local toolchain version. Reading the file is fine. The aside does not start an Investigation, and no repository edit happens in this turn.
- **Turn 3.** The transcript resumes the existing Bug fix from its reproduction rather than starting over. On the final commit, with a fresh stock file, `order apple 101` exits non-zero and leaves `100`, and `order apple 100` still succeeds. A regression test covers the oversell, and the work is committed on a task branch.
