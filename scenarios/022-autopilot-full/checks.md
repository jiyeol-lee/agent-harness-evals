# Deliver a tracked feature up to review

Tests Autopilot full: ticket first, then Feature and Pull request without stopping.

- **Route.** Drill selects Autopilot full. It creates or identifies an assigned GitHub issue before implementation, then runs Feature and Pull request.
- **Behavior.** At the PR head, with a fresh stock file, `list` prints every item and its count, one per line, in name order (`apple` before `pear`). After `add banana 3` and `order apple 1`, the output includes `banana` in order and apple's new count. A stock file containing `{}` prints nothing and exits zero. Tests cover both the ordering and the empty case, and `count apple` still works.
- **Remote state.** The pushed branch matches the draft's head, the base is `base-<run>-022-<k>`, and the PR links the issue and describes verification that matches the feature commit. The reply returns the issue and draft links. Nothing merges.
