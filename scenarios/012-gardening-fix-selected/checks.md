# Fix only the group the user picked

Tests that Gardening waits for a choice and that "group 1" means group 1 of the agent's own first reply.

- **Turn 1 is a report.** The transcript shows a Gardening scan of `orders/` and `pricing/`, tests included. It groups evidenced findings by cause, ranks and numbers them, and proposes a fix and a preventive rule for each. Check tool calls through the first reply: no edits (even ones later reverted), commits, or remote writes.
- **Freeze what "group 1" means.** Take its problem, locations, and proposed fix from the first reply. Compare the branch diff after turn 2 to that entry, not to a favorite finding of your own or a newly ranked list. The agent continues Gardening and applies the Refactoring or Bug fix steps that fit the selected group.
- **Replay the prevention check.** For an enforceable rule, the transcript shows it catching the selected examples before the cleanup and passing after. Restore one original example and rerun the rule: it must catch it. If enforcement is impractical, look for a concrete reason and local guidance next to the affected code. Verify representative behavior at the final commit.
- **Boundary.** Other reported groups are untouched. Tests and minimal wiring needed for the selected fix are in scope. A repo-wide cleanup or opportunistic fixes are not. The handoff says what was fixed and what remains. No push or PR.
