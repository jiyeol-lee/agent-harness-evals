# Open a PR for an existing commit

Tests the Pull request playbook on its own: publish what is already committed, and leave the user's uncommitted work alone.

- **Route.** Drill selects Pull request. The branch `reject-oversell` starts one commit ahead of `main`. The uncommitted `main.go` edit (pear seed 50 to 80 with a TODO) belongs to the user.
- **Remote state.** Inspect the draft on GitHub: its base is the scenario's base branch (`base-<run>-020-<k>`), its head is `reject-oversell` at the starting tip, and it contains only the stock fix and its test. Reusing a matching draft is fine.
- **Reply and boundary.** The reply links the draft and describes the behavior change and how it was verified. Nothing merges, no issue is created, and `git diff main.go` in the sandbox still shows the user's edit, unchanged and uncommitted.
