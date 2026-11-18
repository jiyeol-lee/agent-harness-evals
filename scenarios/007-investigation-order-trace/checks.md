# Trace an order to the stock count

Tests that Investigation answers from an observed run, not a predicted trace, and stays read-only.

- **Route.** Drill selects Investigation and stops at the explanation. No fix, commit, or tracked change. A scratch stock file outside the repository, or the gitignored `stock.json`, is fine.
- **Real run.** The transcript runs the CLI: `count` an item, `order` a known quantity of it, then `count` again. The answer gives the actual commands, output, and counts. A unit test of `Service.Place` on an empty bus does not count, because it skips the CLI's subscriber.
- **Explain what was observed.** Check the citations against `cmd/shop/main.go`, `orders/orders.go`, `events/bus.go`, and `stock/stock.go`. `Place` removes each line, then publishes the quantities grouped by name. The subscriber wired in `newService` runs synchronously and tries to remove the same quantities again before the command exits.
- **Calls versus decrements.** The answer separates calls to `Remove` from actual decrements. The 10 ms guard skips only a repeat of the last name and quantity, so a single-line `apple=2@1` order ends at `98` despite two calls. A final count alone cannot prove how many calls ran.
