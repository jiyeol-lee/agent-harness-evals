# Separate the structural change from the promo

Tests a Refactoring followed by a Feature in one request, with a commit boundary that can be cherry-picked.

- **Refactoring first.** Before moving validation, the agent captures the current results of both `Quote` and `Place` for valid and invalid baskets: empty, nonpositive quantity, and negative price. It runs the same pin against the cleanup. Check the error behavior, and check that an invalid basket does not reserve its valid lines.
- **Commit boundary.** At the cleanup commit, both methods share one validation path, behave as before, and contain no discount-code logic. The pin passes at that revision. The promo arrives only in later commits on the task branch. The final state alone cannot prove this order, so inspect the history.
- **Replay at the feature head.** With a fresh stock file, order `apple=2@5 pear=1@10` with `SAVE10`: the total is `18`. On another fresh file, the same basket with no code totals `20`. On a third fresh file, an unknown code exits non-zero, and `count apple` and `count pear` still print `100` and `50`. The transcript shows these runs, and the feature has a regression test.
- **Scope.** No push, PR, or opportunistic fix to stock, rounding, or the flaky `TestOrderTotal`. Don't grade the flag spelling, helper names, discount representation, or the number of feature commits.
