Autopilot this queue:

1. Add `Stock.Total` to sum the units on hand across all items.
2. Expose it as a `tally total` command, in a second PR on top of the first.
3. Separately, an order for 101 apples when only 100 are in stock must fail without changing the count. Selling the last apple should still work.

Use GitHub Issues in this repo, and {{base}} as the base for the independent changes. Send me three verified draft PRs. I'll merge them myself.
