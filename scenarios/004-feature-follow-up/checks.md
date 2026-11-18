# Extend a feature across a follow-up

Tests the Feature playbook (shape first, real evidence), a requested pause, and that a follow-up continues the same playbook instead of starting a new one.

- **Turn 1, shape first.** Drill selects Feature. Before writing the command, the agent names the output shape and how `list` will read the counts, which `stock.Stock` keeps in an unexported map.
- **Turn 1, evidence.** The reply shows real output from a fresh stock file that decodes to `{"apple":100,"pear":50}`. Ignore key order and whitespace. A test covers `list`, and the code lists the stock's contents rather than hard-coding the seeded names. The agent stops there without the playbook's wrap-up.
- **Turn 2, same playbook.** The transcript continues the Feature playbook. It does not open a new Bug fix or replace `list`.
- **Turn 2, behavior.** On the final commit, with a fresh stock file, `count ghost` exits non-zero. After `order apple 100`, `count apple` prints `0` and exits zero, and `list` includes `"apple": 0`. The work is committed on a task branch.
