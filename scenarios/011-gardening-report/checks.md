# Propose cleanups and wait

Tests that Gardening reports evidence-backed, ranked findings and then stops for the user's choice.

- **Route and scope.** Drill selects Gardening and inspects the named areas locally. The user named directories, so a survey of remote PRs is not needed.
- **Findings.** Verify the location and evidence for each group. Expected examples: validation repeated in `Service.Quote` and `Service.Place`; callers of the deprecated `CalcPrice`; the timed dedupe hack and its `nolint` in `stock.Remove`, next to the second removal path in the `cmd/shop` subscriber; float money; the order-dependent `TestOrderTotal`; linear scans in `Stock`. Credit other supported findings. A suspected bug must be labeled as suspected unless the agent demonstrated it.
- **Decision.** The shortlist groups related symptoms, ranks them by spread and cost, and gives each a concrete fix and a rule that prevents it, such as a regression test for stock effects. It leaves the choice to the user. No source, test, rule, commit, or PR is created.
