# Diagnose before changing the reservation path

Tests a conditional handoff from Investigation to Bug fix: confirm first, then fix the ownership problem rather than the timing symptom.

- **Investigation first.** The transcript shows a read-only Investigation before any edit to `stock.Remove`. Before changing the repository, it ties the observed discrepancy to both callers: `orders.Place` reserves before publishing, and the subscriber in `cmd/shop` removes again. It explains why a single-item order looks fine: the 10 ms guard compares only the last name and quantity. The prompt already authorizes the fix once confirmed, so no second approval is needed.
- **Replay the symptom.** With a fresh stock file, run `order apple=2@1 pear=1@1`, then `count apple`. The original prints `96` instead of `98`. Pear ends at `48` or `49` depending on map order, so don't grade pear on the original. On the fixed commit, expect `98` apples and `49` pears.
- **Fix ownership, not timing.** The task branch has a committed fix with one reservation per order line, and an order still publishes `OrderPlaced` after reserving. Two back-to-back in-process `Place` calls for the same name and quantity both reserve, so the 10 ms guard no longer skips a real order. A regression check covers the duplicate path as the CLI wires it, not just `Place` on a bus without the subscriber.
- **Stay local.** No push, PR, or unrelated oversell redesign. If the agent claims the report is false, inspect its evidence: a passing single-item order does not test the reported case.
