# Concurrent ingestion and Recall diagnostic

Freeze before timing: two repetitions, frontier50/200, four closed-loop readers,
32 measured calls per arm; ordinary and task+lexical mode in reversed order on
repetition2. Same public-fact/hash-embedder/memory-store fixture as read-load.
Two warmups, pack10, adaptive/diversity on. Each arm starts one writer after the
reader gate opens, committing16 records at5ms minimum intervals. Record the
number of active reads observed at each write's start to establish overlap.

Separate modes: in-window records have seed availability (eligible to the fixed
request as-of); future records are available one minute AFTER that as-of.
Track journal insertion attempts and stale rejections through a transparent
store wrapper. Full-frontier event availability is checked via GetEvents AFTER
each timed Recall. No future record may enter a returned frontier. Journal
readback and availability validation are outside individual request timers.

All failures and write durations are retained. Report per-arm p50/p95/max,
attempts/stale rejections and writer overlap. Arm wall time waits for both reads
and writes; ReadWallNS ends when readers finish. Allocations include validation
and writes. Do not compare these allocations directly to read-only Recall cost.
No new latency gate, p99 guarantee or production claim; this is finite in-memory
ingestion, not durable persistence, semantic policy churn or open-loop load.
