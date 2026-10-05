# Immutable graph traversal smoke test

`SearchLayered` implements bounded best-first traversal over immutable layered
records, ef1 at upper levels and a declared ef at level0. It uses a heap for the
candidate queue and a bounded sorted best list. It is not a byte-for-byte port of
the backend's optimized search, nor a completed insertion implementation.

## Observations

The explicit capture test passed execution in1.363s. Quality counts are reported,
not asserted as an exact-search pass. At k10, ef100, metric budget20000:

| N | Snapshot | Queries | Exact top10 hits | Self misses | Max metric evaluations |
| --- | --- | ---: | ---: | ---: | ---: |
| 800 | initial | 8 | 80/80 | 0 | 908 |
| 800 | after mutations | 8 | 80/80 | 0 | 893 |
| 6400 | initial | 8 | 79/80 | 0 | 3016 |
| 6400 | after mutations | 8 | 78/80 | 0 | 2998 |

Combined317/320 expected neighbors. The queries are vectors already stored in the
graph, selected at evenly spaced sorted ordinals. This is deliberately an easy
smoke test; it cannot support general/OOD retrieval, calibration or agent claims.
The exact oracle exhaustively scores all vectors outside the search algorithm.

Three race runs of dimension/budget/self-search controls passed in1.308s. Budget
exhaustion returns no partial result. The capture test itself ran without race.

## Cost and limits

An ef100 retained candidate list does not imply100 evaluated records. The observed
maximum is3016 evaluations across levels. Best-list insertion shifts O(ef) entries;
this baseline has not been optimized or latency-benchmarked. The evaluation cap
does not separately limit all link reads, missing-node visits, queue/seen-map bytes
or copied output strings. No total RAM or end-to-end deadline guarantee follows.

Next add held-out query controls and reuse tested traversal in insertion neighbor
discovery, with the backend's selection and connection behavior verified explicitly.
Three missing exact neighbors remain honest approximate-search errors, not hidden
by the successful test process. All seven whole goals remain open; production and
standard dependencies are unchanged.
