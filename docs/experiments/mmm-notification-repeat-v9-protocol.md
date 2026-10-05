# Notification load replication v9

Frozen 2026-10-01 before this repeat. Re-run the existing
`TestResearchLearningNotificationComparison` without changing its source,
workload, thresholds, worker, queue size, store adapter, or input fixture.
This is the same 36-arm, six-trial, 192-recall/96-future-write experiment as
v8: queue capacities 16 and 64, with rotated off/notified/polling order and
50 delivered labels per admitted frontier. The fresh process and temporary
stores supply a scheduling replication, not a new task distribution.

Retain the v8 frozen per-cell requirements: no read/write/bridge/worker errors,
read/write overlap, at least 154/192 admitted frontiers, completion of all
50 labels per admitted frontier, no worker failures, completion age p95 at
most 250 ms, and enabled serving p99 at most 1.10 times paired off p99.
Report all 36 arms and all failures. Queue-64 is a replication candidate;
queue-16 remains a recorded v8 partial failure. Do not select a favorable
trial, round away a boundary miss, or change production settings.

The test writes an exclusive-create source-hashed JSONL artifact. Independently
recompute sample counts, nearest-rank quantiles, label conservation and gates
from that artifact. The result can establish repeatability only for this finite
synthetic arrival pattern. It cannot establish population p99, realistic mixed
mutations, durable recovery, prediction quality, or agent-task benefit.
