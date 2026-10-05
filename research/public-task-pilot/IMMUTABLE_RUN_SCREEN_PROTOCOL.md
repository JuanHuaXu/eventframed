# Immutable run construction/read tradeoff screen

Freeze before execution. Two independent graph builds,800 initial records,
768-dimensional SHA256-derived vectors, CPU cap4. Eight stages each contain32
mutations:4 updates and2 deletes of distinct original IDs,26 new IDs. Compare
building only that stage's run (retaining up to9 runs) against rebuilding the
entire current state. Alternate construction order by stage/repeat; alternate
query order per probe. Do not run unrelated tools during measurements.

Each stage uses32 independent deterministic probes and exhaustive current-state
cosine top10. Persist candidates, oracle and per-query durations, graph build
and plan construction times, plus source hashes. Run append build+plan must be
less than full rebuild+plan in both repetitions at every stage to count as a
consistent construction benefit. Mean recall at each stage must be at least.99
and no more than.005 below full rebuild. Every successful query must finish
under100ms; stale or deleted candidates and errors fail. Report query slowdown.

This is a static feasibility screen, not a load/capacity proof: the finite run cap
eventually requires compaction. There is no authoritative transaction workload,
concurrent writes, retirement pressure or included final compaction debt here.
Graph closes are recorded separately. Passing cannot complete goal6.
