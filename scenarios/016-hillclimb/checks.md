# Tune order latency against a fixed budget

Tests Hillclimb: a frozen metric, one commit per kept change, and an honest stop.

- **Route and workload.** Drill selects Hillclimb. The harness calls `Service.Place` with the `cmd/shop` subscriber attached (or the same wiring), over 100,000 distinct stocked names. It is not the two seeded names, one order with 100,000 lines, or a `Count` microbenchmark. Inspect how it picks names across the catalog, provisions enough stock, and counts completed orders and errors.
- **No fake wins.** Repeating the same name and quantity within 10 ms makes `Remove` skip the reservation, and a fast error is not a successful order. Either one is a measurement bug, not a gain. Setup is outside request timing.
- **Metric.** Before changing the slow path, the agent freezes a repeatable command, records a correctness baseline and repeated p95 measurements, and sets the stop check. Look for the median of run-level p95s in microseconds, not mean latency.
- **Attempts.** For each attempted change, the decision log runs from mechanism through measurement and regression check to keep or revert. Kept improvements beat noise without breaking behavior, each in its own local commit. Reverted experiments are gone from the code. No push or PR.
- **Stop honestly.** Compare the final result to the same strict `< 5 µs` budget and the stated stop check. If time or noise prevents a conclusion, credit accurate evidence and limits, not a claim that the target was met.
