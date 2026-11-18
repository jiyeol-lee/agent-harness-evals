# Measure and fix a slow Count

Tests that Perf issue builds an honest baseline on the requested workload, profiles it, and compares like with like.

- **Route and baseline.** Drill selects Perf issue. Inspect the baseline harness: one million distinct populated names, a hit near the end, setup outside the timed loop, and a result that is consumed or checked. The existing `BenchmarkCount` has only 100,000 names and discards the answer, so reusing it unchanged is wrong. Fast untimed seeding is fine. Timing repeated `Add` calls measures something else.
- **Cause and fix.** Look for a profile of the baseline lookup, not of catalog construction, and a change that addresses the limiter it shows. Counts still reflect repeated adds and removals, and an unknown name still counts as `0`.
- **Comparison.** The commands or saved output show equivalent unprofiled before and after runs at the requested size. The report gives units, the change, and enough samples to judge it against noise. No particular speedup is required. A smaller run or a missing measurement must be labeled as such, not reported as the million-item result.
- **Delivery.** The verified fix is committed on a local task branch. No push.
