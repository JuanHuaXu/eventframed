# Grouped durable admission v29

Frozen before execution. Compare off, group4 without persistence, and group4
with the actual exclusively owned Durable SQLite wrapper, three rotated trials.
Retain 192 recalls, four readers, 96 future writes at 2ms intervals, 50 candidates,
four-observation ready groups, 64-slot handoff and 20ms entry deadline. Fresh
LibraVDB and SQLite paths per arm. No labels, fitting, retries or production use.

After all group members validate under the existing as-of mutation guard, admit
each candidate with its service binding through Durable.AdmitBound, read and
compare the entire persisted record with its cold preview, then Durable.Discard.
These are real individual SQLite transactions and readback operations, not a
simulated persistence delay. Both writes and integrity readback are included in
callback time and accepted age. Admission is counted only after the whole group
completes. Any partial durable error is fatal and preserved, never retried as
though the group were atomic. This is NOT a cross-store transaction.

Preview IDs remain local. Durable IDs are contiguous over admitted records only,
so a pre-admission deadline does not create a gap. Rebinding is performed on a
detached preview before validation; all original fields must exactly match the
real durable record after admission. This equivalence relies on the deliberately
cold, unlabeled fixture and must NOT be reused for a concurrently learning model.

Record every attempted group, weighted outcomes, raw request/write/guard/age
durations and actual admit/discard counts. Require admits=discards=validated=
50*accepted in successful arms; no stored discard is evidence. Small read-only
and write arms run three times under race first. Then exclusive-create nine-arm
JSONL with source hashes, dependency/runtime metadata and per-arm fsync.

Retain the 250ms age and 1.10 paired read-p99 screens and report write-cost
transfer, drops and deadlines. Do not claim durable learning, recovery under
concurrent failure, or real-agent accuracy from this storage-cost experiment.
