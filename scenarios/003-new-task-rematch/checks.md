# Switch playbooks when the user says "new task"

Tests that a finished playbook ends cleanly and that "New task:" makes Drill match again instead of continuing the old playbook.

- **Turn 1.** Drill selects Bug fix and finishes it before the second message. The reply reports a verified fix: `order apple 101` exits non-zero, `100` remain, and a regression test covers it. The fix is committed on a task branch.
- **Turn 2.** The transcript shows a fresh match to Feature for `Stock.Total()`, not another step of the oversell Bug fix. `Total` returns `0` for an empty stock, `5` after adding three apples and two pears, and `4` after removing one pear. Run `go test ./...` on the final commit: the oversell regression test still passes alongside the feature.
