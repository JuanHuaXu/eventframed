# Durable transaction comparison v33

Frozen before execution. Twelve rotated arms: off, group4 without persistence,
group4 individual transactions, group4 batch transactions; three trials each.
Retain 192 recalls, four readers, 96 future writes at 2ms intervals, 50 candidates,
four-observation ready groups, 64-slot handoff and 20ms entry deadline. Fresh
LibraVDB/SQLite paths per arm. No labels, fitting, retries or production changes.

All group observations validate under the as-of guard first. The new arm uses
one AdmitBatch and one DiscardBatch transaction for all up to 200 candidates in
that group. Actual owned-worker original records must exactly match cold fixture
previews, and every persisted record is still read back and compared before
discard. No integrity check or FULL synchronous durability setting is removed.
This does not generalize preview-ID relabeling to a trained concurrent learner.

Actual admission/discard operation counts remain PER CANDIDATE. Group phase
timings include real batch APIs and complete readback. Partial failures are
fatal and preserved, never disguised as an atomic cross-store rollback. No
feedback is inferred from discard. Accounting must conserve outcomes, group
sizes, records and monotonic phase containment.

Run small persistent read-only/write arms under race three times before the
full non-race comparison. Preserve source hashes, runtime/dependencies and every
raw timing in exclusive-create JSONL. Report nearest-rank read/write p99, age
p95 and operation counts. Retain 250ms age and 1.10 paired read-p99 screens;
report write-cost transfer separately. A finite storage-cost rescue is not a
full durable learning, lifecycle, provenance or real-agent accuracy result.
