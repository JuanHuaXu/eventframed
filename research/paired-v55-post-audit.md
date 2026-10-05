# V55 post-audit repair

The first checkpoint attempt terminated1 BEFORE creating its output directory.
It compared failure.json's error to the bare string `vet`. Node's assertion
message actually contains `vet` followed by its formatted expected/actual
details. Confirmed archive-helper classification bug, not model or timing data.
The preserved structured command record names vet and exitCode1. Verify those
fields instead; do not alter original failure text, logs, source freeze or
scientific outcomes. The next checkpoint attempt uses the same still-absent
exclusive output directory; no live run is restarted and no artifact replaced.
