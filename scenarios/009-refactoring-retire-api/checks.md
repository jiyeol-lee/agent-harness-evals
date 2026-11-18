# Retire a deprecated API

Tests that Refactoring pins current behavior before moving callers and that the migration is complete.

- **Route and baseline.** Drill selects Refactoring. Before editing call sites, the agent records current pricing behavior with a characterization test, snapshot, or runnable comparison. A search or a clean build does not pin behavior.
- **Complete migration.** Check `orders/orders.go` (both `Quote` and `Place`), the `price` command in `cmd/shop/main.go`, and `orders/orders_test.go`. All work without `CalcPrice`, and a search of the final source and tests finds no reference. Pricing goes through `pricing.Quote` with no discount. There is no compatibility wrapper or second pricing implementation.
- **Equivalence.** The same cases give the same results before and after, including a multi-line order and a real `price` run: `go run ./cmd/shop price 2.5 2` prints `5`. Validation and reservation behavior are not redesigned.
- **Delivery.** The migration and the deletion are committed together on a local task branch. No push.
