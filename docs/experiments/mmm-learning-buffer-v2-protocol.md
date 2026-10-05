# Bounded-buffer rescue v2

Freeze before execution. Preserve v1 failures. Compare off, tap16, tap64 in
rotated order over three trials, each with fresh persistent stores. Repeat at
64 and192 read requests (four readers), with32 and96 future writes respectively.
All other learning/feedback contracts and50-candidate frontiers remain unchanged.
192 journals remains within the existing256 replay-ledger bound.

Record every Recall/write latency, admission/drop/error count, completed labels,
and enqueue-to-completed-update duration for each admitted frontier. Timestamp is
at the tap's nonblocking enqueue attempt; it excludes prior Recall assembly.
Admitted work still must pass the temporal proof; larger buffers do not relax it.

For each on arm, trial and workload, require: zero errors, overlap>0, >=80% of
offered frontiers completed, all admitted labels completed, paired p99 <=1.10*off,
and completion-age p95 <=250ms. The extra age gate prevents a larger backlog
from masquerading as fast learning. Nearest-rank percentiles. Queue64 must pass
both workloads, not merely the64-request burst, to advance. No steady-state or
population guarantee follows even if this finite screen passes.

Report structural buffer size separately: at most64 observations of50 compact
candidate records in this fixture, versus16. This excludes the worker, bridge
ledger and store. No claim of total process-memory reduction is intended.
Exclusive raw artifact creation before runs, append and sync each finished arm.
