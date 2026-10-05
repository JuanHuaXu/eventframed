# Durable component cost v10

All12 diagnostic arms completed and all512-state restart comparisons passed.
Raw measurements/source snapshots: mmm-durable-cost-v10.jsonl. Independent audit
verified source hashes/current files, unique arms and latency-array lengths.
Package vet passed. No performance success gate was declared for this diagnostic.

| Labels | Memory total ms | Durable total ms | Full replay ms | Closed DB bytes |
| --- | --- | --- | --- | --- |
| 256 | 10.844-22.363 | 40.758-46.169 | 11.891-12.324 | 241664 |
| 1024 | 43.670-43.873 | 144.843-149.883 | 46.394-47.771 | 933888 |

At1024 labels, paired total-time ratios are3.416/3.317/3.328. Durable admission
p99 is153-190us and feedback p99 is117-152us across those three trials. Total
includes waiting every64 labels for actual completion, not only enqueue time.
Database size is measured after close, not peak WAL footprint or memory use.

This shows a measurable durability cost on the current local environment, with
small absolute component timings. It does not establish physical power-loss
guarantees, remote-disk performance, concurrent retrieval latency, steady-state
throughput under other arrival patterns or a universal sub100ms bound. Replay
and database growth across these sizes make checkpoint/retention research
relevant; two sizes are not an asymptotic proof or a measured long-history bound.

Next integrate the durable consumer off the serving path and measure loaded
serving plus queue age, while preserving original forecasts and dependency
authority. Do not compare different modes' final forecasts as accuracy evidence:
their asynchronous admission timing can legitimately differ. Each replay was
compared only against its own pre-close model.
