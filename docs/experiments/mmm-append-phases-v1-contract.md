# Prepared append phase attribution

Isolated research SQLite ledger, source-identity index enabled, WAL and
synchronous FULL unchanged. Reuse unique-source1KiB-padding fixtures; batches
50/200,32 batches per cell,3 trials with measured/unmeasured order rotated.
No production DB, schema, source owner or publication guard is changed.

The test-only measured copy preserves AppendBatchPrepared's statements,
validation, retries, atomicity and commit semantics. Time preflight validation
and identity serialization, BeginTx, statement preparation, row work, and Commit.
Total includes deferred cleanup; phases must not overlap or exceed total.
The unmeasured original is the control; never infer zero instrumentation effect.

Payload construction is outside append timing in both arms. These measurements
exclude source-owner resolution, worker prediction, original payload marshaling,
publication-gate acquisition and foreground queueing. They can locate append
cost, not prove full-service latency or bound queue tails by simple subtraction.

Parity tests: exact fresh/retry acknowledgments, canceled precommit rollback,
source-identity conflict rollback and persisted rows. Reopen every experimental
ledger and verify every key/source/payload. Capture source snapshots and all384
calls across12 cells. Preserve all trials, no threshold or schema tuning.
