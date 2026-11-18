# Hand back an auditable cents migration

Tests that Drill sends cross-cutting work the user reviews later to `figure-it-out`, and that the result is auditable.

- **Route.** From Drill, the agent loads `figure-it-out` instead of a single Feature, Refactoring, or Autopilot playbook. Before production edits, look for a written workflow, a pass/fail definition of done, the risky contract decisions, and baseline cases captured for comparison.
- **Contract.** Follow a cent amount from `shop price` and `shop order` through the order and pricing types to the output. Money is stored and computed as integer cents, not just converted when printed. Check the documented units, input validation, and the chosen fractional-cent discount rule against examples. No particular rounding direction or flag spelling is required.
- **Proof.** Inspect tests and real CLI runs: a 10-cent item plus a 20-cent item total exactly 30 cents, where the original prints `total 0.30000000000000004` for `order apple=1@0.1 pear=1@0.2` on a fresh stock file. A unit price times quantity gives the expected cents through `shop price`. A discount that yields a fractional cent follows the stated rule. Invalid quantities are still rejected, and valid orders still reserve stock.
- **Review tomorrow.** The handoff points to a durable decision log. It links independently reviewable local commits to their hypothesis, check, and keep or revert result, and separates verified outcomes from open risks. A plan alone is not a finished migration. The sandbox has no remote, so a claimed PR is a fail.
