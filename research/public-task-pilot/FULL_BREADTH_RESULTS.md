# Full-breadth diagnosis: reachable, but costly

Protocol: FULL_BREADTH_PROTOCOL.md. Artifact: `full-breadth-results.json`.
Two fresh6400-record768d graphs,1024 self queries per graph, two search arms.
The direct stored-vector check is independent of the in-memory reference map.

| Effort | Misses build0 / build1 | Median ms build0 / build1 |
|---|---|---|
|Default (effective400)|1 / 1|0.947 / 0.952|
|6400|0 / 0|3.725 / 3.751|

The default miss is seed-954 in both runs. Every queried record, including that
target, is present with exactly the expected vector in the derived database.
All4096 queries complete without search errors; default results match the
existing adapter. Independent verifier checks hashes and recomputes every
returned candidate cosine, counts and hit flags.

The full-breadth query retrieves seed-954, establishing that it is reachable in
these two new graphs. Missing durable payload or a permanently unreachable target
does not explain these particular default misses. This is not an exact diagnosis
of the older growth graphs, which were not reopened or altered. It also does not
prove every graph node reachable or every nonsynthetic query correct.

Median query cost rises3.93x/3.94x. Query breadth equals corpus size here, so this
is a costly diagnostic bound, not the desired bounded-frontier serving design.
Do not adopt it as a universal fallback or call two discovery builds a validated
rescue. Storage reads precede both arms, warming payloads; timings are unloaded
and are not directly comparable with previous load studies.

Remaining routes: improve graph construction/entry coverage at fixed query work,
or compare a bounded-fanout immutable partition design with exact-oracle recall
and compaction-inclusive load. The latter must handle version shadowing across
runs and cap retired-reader resources; reducing rebuild size must not silently
turn retrieval into an unbounded full scan. No production changes. All seven
research goals remain open.
