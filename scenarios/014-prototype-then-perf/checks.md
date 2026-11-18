# Choose a Count implementation by experiment

Tests a Prototype that decides a design, followed by a measured production change.

- **Prototype first.** Tool calls show both a sorted, binary-search lookup and a name-keyed map built and run in scratch space, outside the shop's tree. Both use the same 100,000-item catalog and look up `item-99999`. Setup is outside the timed lookup, and results are consumed or otherwise shown not to be optimized away. The agent records the scratch path and comparable numbers for each.
- **Decision before production edits.** The agent picks one using its measurements and a concrete tradeoff, such as insertion cost, memory, or keeping a second representation in sync. It then moves from Prototype to Perf issue or Feature work in the shop. Either candidate can win.
- **Validate the real implementation.** The transcript has comparable timings for the original and final production `Stock.Count`, not only the scratch numbers. Replay the command. Adds to an existing name accumulate, a removal changes the count, and an unknown name counts as `0`.
- **Scope.** Only the chosen implementation and its verification are committed on the task branch. No abandoned variant, scratch project, made-up speedup, or push. Grade the evidence and the decision, not machine-specific timings.
