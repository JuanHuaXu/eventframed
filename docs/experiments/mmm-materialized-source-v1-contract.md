# Materialized source-table comparison

Admission-only test-owned schema, no production migration. Compare original
AppendBatchPrepared plus service-identity index against materializedSourceAppend,
including canonical validation inside measured append time. Same fixture and
WAL/FULL durability; batch50/200,32batches,3 rotated trials. Verify every original
after reopen by source/key/sequence/payload and time those point reads including
source-to-payload validation. Record every sample, no exclusions or tuning.

Correctness precondition: atomic duplicate-source rejection (including equivalent
JSON escapes), exact retries, changed-byte rejection, cancellation rollback,
reopen persistence, and corruption rejection. Private candidate writer only;
raw SQL can bypass source-to-payload derivation and is not authorized.

This is not production-equivalent: no feedback records, migration, process-crash,
concurrent ownership, corrupted-row allocation bound, or complete legacy Unicode
compatibility claim. A speedup on admissions alone cannot complete goal6.
