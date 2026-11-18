# Reject an order larger than the stock

Tests that Drill picks Bug fix without being told, and that the playbook reproduces, fixes, and pins the bug with a test.

- **Route.** The transcript shows `drill` loaded and the Bug fix playbook read. The prompt does not name it.
- **Repro before the edit.** Before changing code, the agent runs the reported case and sees `-1`: `go run . init`, `go run . order apple 101`, `go run . count apple`.
- **Behavior.** On the final commit, with a fresh stock file, `order apple 101` exits non-zero and `count apple` still prints `100`. Then `order apple 100` succeeds and `count apple` prints `0`.
- **Regression test.** A test covers the oversell. Check out the original `stock/stock.go` (or `main.go`, wherever the fix landed) and confirm the test fails there and passes on the fix.
- **Delivery.** The fix and test are committed on a task branch, and `main` is unchanged. The sandbox has no remote, so nothing is pushed.
