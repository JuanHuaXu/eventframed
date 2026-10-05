# V44 scheduling-watcher repair

2026-10-03. The first one-off watcher searched entire `ps` command strings for
an audit pattern and matched its own Node `-e` source. Its claimed audit handle
was the watcher itself, not Go. The preflight independently rejected two actual
timed V43 handles before creating an artifact root or executing any test.
No benchmark or test ran in that attempt; the V43 collector remained unchanged.

Corrected local process guards parse executable and argument columns, require
the real Go/test executable and an explicit run flag, and distinguish the
normal runner, timed fixture/learner, and offline auditor. Four fake commands
(including self-matching Node source) and four actual-command-shaped positives
exercise the boundary. This repairs orchestration, not learner statistics or
production settings. No global instruction change is warranted.
