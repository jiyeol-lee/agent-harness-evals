# Compare two designs by running them

Tests that a "which design?" question that an experiment can settle goes to a throwaway Prototype instead of a tradeoff table.

- **Route and isolation.** Drill selects Prototype, not an Investigation that stops at a tradeoff table. The transcript names a throwaway prototype under `/tmp/agentic-coding-tool/`, outside the checkout. No production change, commit, or PR, and `git status` in the sandbox is clean.
- **Experiment.** Inspect the scratch code and its run output. One variant sums quantities on read, the other keeps a running total. Both get identical inventory, reads, adds, and removals. The observed totals after each step match the expected quantities. Note that the shop's `Remove` has a 10 ms dedupe: a prototype that reuses it must account for it.
- **Decision.** The recommendation uses the run to weigh read cost against update bookkeeping or stale totals. Either choice can pass. A behavioral comparison need not claim a speedup. If it does, look for real, comparable timings.
