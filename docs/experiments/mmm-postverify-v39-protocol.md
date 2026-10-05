# Post-guard original readback v39

Frozen before execution. Twelve rotated cells, three trials of off, non-durable
group4, post-guard discard with readback inside, and post-guard readback plus
discard. Keep the v38 fixture: 192 recalls/four readers, 96 future writes spaced
2ms, 50 candidates, ready groups capped at four, handoff 64, 20ms entry deadline,
fresh persistent stores and FULL SQLite durability. No labels/fits/production.

Service validation, actual owned-worker admission and comparison with its returned
original remain inside the guard. Only the immutable ledger reread joins typed
discard outside. All original fields must compare exactly. A missing/mismatched
record, cancellation or database error prevents discard and completion. Finish
the whole group before taking another; do not expand pending capacity. Releasing
the guard conveys no fresh service, label or model-history authority.

Age and total duration include both post-guard operations. Count completion only
after successful verification and terminal commit; report partial failures as
errors, never as stale/expired admission or completed learning. Keep source hashes,
all raw timings, group-weighted counters and actual durable operation counts in
exclusive-create JSONL. Run publication/reopen, failed-read/no-discard, small-load
race tests plus full ledger/learner/service race and vet before the full run.

Screens remain 250ms accepted-age p95 and read-p99/off <= 1.10. Report writer
tails, drops and expiries independently. All earlier failures remain. This can
improve overlap without eliminating durable cost; failure to pass age remains
failure even if guard duration shrinks. No learned-model/real-agent conclusion.
