# Watch CI on an existing draft

Tests the Autonomous run playbook: re-check an outside condition (remote CI) after each change, and stop only on the new head's result.

- **Setup, not agent work.** `setup.sh` commits a refactor of `Remove` that never stores the new count, and the runner opens a draft for it against `base-<run>-024-<k>`. At the seeded head, `go vet` passes and `go test` fails `TestRemove` with `Count(apple) = 10, want 7`. GitHub Actions must be enabled. No check runs, or CI that never starts, makes the run inconclusive, not green.
- **Diagnose from CI.** Drill selects Autonomous run. The agent finds the existing PR, reads the failed check and its log, and reproduces the failure locally. A local test alone cannot establish the remote check state.
- **Minimal fix on the same PR.** The PR gets a small fix, such as storing `left` back into the map. The original tests pass without edits to their assertions, the workflow, or branch protection. No opportunistic oversell fix.
- **Wait for the new head.** Inspect check runs for the new head SHA. The agent waits for them to finish and stops on success, not on an earlier head's result. The iteration log, verification, and handoff match the PR. No duplicate PR, merge, or deploy.
