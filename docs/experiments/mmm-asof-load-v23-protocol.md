# As-of guard load v23

Frozen before execution. Nine rotated arms: off, queued exact validation and
queued as-of validation, three trials each. Match v22: fresh persistent LibraVDB,
192 recalls, four readers, 96 future writes spaced 2ms, 50 visible candidates,
64-slot handoff. Both enabled modes have 20ms entry deadlines, no retries,
full 50-candidate validation, no ledger writes and no labels or fitting.

The only intended guard difference is temporal acceptance: as-of mode requires
committed native/publication agreement and complete future-only owned ingestion
history relative to the fixture query time. Backfills, policy changes, missing
history and quarantine remain invalid. The request and record retain that query
time. Callback I/O uses the enclosing one-minute context, not the entry deadline.

First test small read-only and concurrent-write cases under race. Then record
all nine non-race arms, selected sources/hashes, runtime/dependencies and raw
request/write/guard/age timings in an exclusive-create, per-arm-fsynced JSONL.
Conserve attempts+drops=192 and attempts=accepted+busy+stale+expired;
validated=50*accepted. Report accepted counts, drops, errors and nearest-rank
read p99, write p99, guard p95 and accepted-age p95.

This diagnoses whether correct temporal admission solves the stale barrier.
It does not establish learning or durable performance. Prior 250ms age and
1.10 paired recall-p99-ratio targets remain screening references; do not call
admission alone an overall rescue or substitute a new adoption threshold.
