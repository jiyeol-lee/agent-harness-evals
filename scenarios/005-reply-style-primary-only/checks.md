# Load the reply style in the primary agent only

Tests that the primary agent loads `output-styles` before it replies and that subagents never load it.

- **Primary loads it.** In the transcript, the primary agent calls the `skill` tool with `output-styles` before the text of its final reply. A reply written before any `output-styles` load fails.
- **Subagents never load it.** The transcript shows at least one `subagent` call. Without one, record inconclusive. Export each child session named in the `subagent` tool output with `opencode session export <sessionID>`. No child session calls the `skill` tool with `output-styles`, including an attempt that the permission denies.
- **Answer.** The summary names the `order` and `discard` cases in `main.go` and the tests in `stock/stock_test.go`, with file paths. `git status` shows no tracked changes.
