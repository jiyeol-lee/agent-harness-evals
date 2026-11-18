# Explain the removal shortcut without fixing it

Tests that a how-does-this-work question routes straight to Investigation and stays read-only.

- **Route and scope.** Drill selects Investigation directly. The agent reads `stock/stock.go` and cites the relevant lines. `git status` shows no tracked changes, tests included.
- **Explanation.** The answer follows two consecutive calls. The first removal of an unknown name records the attempt in `last` before the lookup and returns `ErrUnknownItem`. A second call with the same name and quantity less than 10 ms later returns nil before any lookup or subtraction. So yes, it can happen for a never-stocked name. The shortcut compares only the last recorded attempt, not a history of successful removals.
