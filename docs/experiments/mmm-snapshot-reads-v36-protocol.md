# Snapshot-read integration v36

Frozen before execution. Twelve rotated arms: off, non-durable group4, batch
transactions with original per-key reads, and batch transactions with snapshot
reads for admission/discard preflight and complete original-record readback.
Three trials each. Retain 192 recalls, four readers, 96 future writes at 2ms,
50 candidates, four-observation ready groups, 64-slot handoff, 20ms entry deadline
and fresh stores. No labels, fitting, retries or production configuration changes.

Both durable arms use the same validated write-batch semantics and FULL
synchronous SQLite. Every original is still checked against the owned-worker
record and reread from the database. Only read grouping changes. The experimental
wrapper validates lookup response length, exact keys/kinds, missing-row shape,
payload bounds and canonical prediction identity before use. Missing admissions
fail full readback. Errors during preflight cannot stage learner state.

Read primitives bundle transaction and statement reuse; this experiment does
not distinguish their individual effects. All prior failures remain. Run warm
parity/malformed-response/recovery tests and small persistent load under race
before the full non-race comparison. Preserve all raw timings, source hashes,
group accounting and actual durable operation counts in exclusive-create JSONL.

Retain 250ms accepted-age and 1.10 paired read-p99 screens; report write tails,
drops and expiries. No loaded warm-learning, provenance authority or real-agent
claim follows from these cold, unlabeled storage fixtures.
