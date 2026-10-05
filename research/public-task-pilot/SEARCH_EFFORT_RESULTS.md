# Search-effort diagnostic: no robust rescue

Artifacts: `search-effort-results.json` and `search-effort-v2-results.json`.
Protocols: SEARCH_EFFORT_PROTOCOL.md and SEARCH_EFFORT_V2_PROTOCOL.md.
Both retain full candidate traces from two fresh6400-record/768d graphs.

The first experiment mistakenly chose requested efforts0/100/200/400 below or
at the backend's effective floor400. All missed seed-954 in both graphs. This
is retained as an equal-effective-effort control, not an increased-effort test.
The floor is explicit in pinned libravdb internal/index/hnsw/hnsw.go:
`max(EfSearch, k, 2*EfConstruction, efOverride)`.

The corrected experiment uses effective400/800/1600/3200 on each SAME graph.

| Effective effort | Misses, build0 / build1 (1024 each) | Median ms, build0 / build1 |
|---|---|---|
|400 (default)|1 / 1|0.899 / 0.913|
|800|1 / 1|1.156 / 1.178|
|1600|1 / 1|1.591 / 1.617|
|3200|1 / 0|2.454 / 2.491|

Every miss is seed-954. All16384 recorded queries across both experiments have
no search error; the4096 default results exactly match the existing empty-delta
adapter controls. Independent verification recomputes every candidate cosine
and checks source hashes, IDs, counts and hit flags.

Increasing effort to3200 raises median query time about2.73x while recovering
the target in only one graph. This fails a robust rescue criterion even on the
known failing fixture. No serving defaults are changed. Unloaded durations are
not tail-latency evidence under concurrent writes. The two graph builds share
the deterministic vector dataset; they are not independent task samples.

The backend accepts the override and timing increases, but these results do not
identify its internal failure mechanism. Next check indexed-record presence and
graph reachability, or compare a separately validated index construction. Treat
the throughput problem separately: larger effort does not reduce rebuild cost.
All seven research goals remain open.
