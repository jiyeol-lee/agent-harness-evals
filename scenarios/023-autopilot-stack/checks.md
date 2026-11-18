# Stack only the dependent change

Tests Autopilot stack: a queue mixing Feature and Bug fix work, stacked only where one task builds on another, each head independently reviewed.

- **Route.** Drill selects Autopilot stack and runs Autopilot full for each task: Feature for tasks 1 and 2, Bug fix for task 3. Issues, branch tips, and draft PRs refer to the same changes.
- **Bases and diffs.** On GitHub, the `Total` PR and the oversell PR each target `base-<run>-023-<k>`. The `tally total` PR targets the `Total` branch and adds only the command, not a second `Total`. The oversell PR contains neither `Total` nor the command, and the `Total` PR does not contain the oversell fix.
- **Replay each head.** At the `Total` head, `Total` returns `150` for a fresh stock file and follows adds and removals. At the command head, `tally total` agrees with that after `order pear 10`. At the oversell head, with a fresh stock file, `order apple 101` exits non-zero and leaves `100`, and `order apple 100` leaves `0`. Each head has tests for its own change.
- **Review and handoff.** Each delivered head has one independent review with evidence and a verdict, by a reviewer on the other model from the one that wrote it. The handoff names each draft, its base and head, and open problems. Nothing merges, and temporary verification worktrees are removed.
