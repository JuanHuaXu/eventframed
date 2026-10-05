# Fixed partition screen, before serving integration

Compare one base with eight ID-hash-owned immutable partitions,6400 records,
768d,CPU4,two builds of each layout. Every query visits all partitions sequentially
and merges local top10. HNSW parameters unchanged; no corpus-sized breadth.
Total query work is higher than one base: effective400 per partition, not400 total.

Use1024 self queries plus128 independent hash-vector probes. Compute exact top10
by exhaustive cosine for probes before measured queries. Alternate layout query
order. Retain local candidates, merged candidates, oracle results and durations.
Measure each initial build and an unpublished replacement of partition0 with32
new hash vectors routed to that partition. No durability, concurrent mutation or
serving publication is tested here. Original derived stores remain untouched.

Screen gates for eight partitions: no self miss/error; mean probe recall>=0.99
and no reduction exceeding0.005 against same-repeat single-base control; maximum
unloaded query<100ms; replacement build<320ms (the earlier32-slot headroom at
100 writes/sec). These are necessary discovery gates, not sufficient load proof.
Passing does not show billion-event scaling: fixed partition count leaves shard
size growing with corpus; increasing count requires migration and changes cost.
Do not raise partitions, lower load, or relax recall after observing results.
