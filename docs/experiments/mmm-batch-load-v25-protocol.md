# Bounded batch validation v25

Frozen before execution, after the v24 profile. Nine rotated arms: off, original
per-record as-of validation, bounded batch as-of validation; three trials each.
Each uses fresh persistent LibraVDB, 192 recalls, four readers, 96 future writes
spaced 2ms, 50 visible candidates and a 64-slot handoff. No feedback, fitting,
ledger writes, retries or production configuration changes.

Both enabled arms retain the same 20ms entry guard and final dependency check.
Batch mode validates every original forecast against actual journal membership,
baseline, event features, query, tenant, snapshot and as-of time. It reads the
shared journal/query once and fetches the frontier events in one API call. Only
after ALL records pass does the callback count them as validated. Duplicate
event/prediction IDs, mixed bindings and batches outside 1..256 reject. Single
validation remains unchanged as an independent control. Neither mode learns
from discard or invents labels. Temporal and durable authority gaps remain.

First run multi-record parity/negative controls and small read-only/writer load
under race. Then collect raw nine-arm JSONL with selected source hashes and
runtime metadata. Conserve attempts+drops=192 and attempts=accepted+busy+stale+
expired; validated=50*accepted. Report nearest-rank read/write p99, guard p95,
accepted-age p95 and admission/drop counts. Entry deadline does not bound a
callback ignoring cancellation; callback I/O uses the enclosing minute context.

Retain prior 250ms accepted-age and 1.10 paired read-p99 ratio screens. A rescue
requires timely admission without merely transferring costs to writes or
discarding work. This is infrastructure evidence, not predictive improvement;
do not promote finite performance success into MMM accuracy validation.
