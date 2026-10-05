# Incremental diversity integration: finite overload rescue

## Exact service replay

The fresh incremental-overlay-v1 changes only the diversity call inside packing
Select, retaining the prior task-lexical service overlay. Original source files
and older artifacts are unchanged. Both the ordinary and experimental packing
arms receive the same optimization.

All108 outputs on the five public datasets equal their previously recorded
outputs after removing duration only. Comparison includes selected candidates,
scores, laws, support ranks, rank traces, confidence fields and parsed journal
explanations. Existing incorrect answers and eight absent-answer cases remain;
this is a performance change, not new learning or higher accuracy.

The complete packing and researchcalendar test packages pass under the overlay
with the race detector. The original diversify helper remains available as the
reference in exact-output tests.

## Fixed-arrival results

All32 arms completed:1024 reads and512 writes. Every one of the16 admitted arms
passes the frozen no-error/no-stale/100ms screen, including the previously failed
200-read/s+100-write/s conditions. The admitted workload totals512 reads and256
writes with zero errors, zero stale journal rejections, zero observed future
evidence and matching journals on successful reads. Largest admitted read:
20.136ms; largest admitted write:11.487ms, including scheduling/admission wait.

| High-rate admitted aggregate | Previous packing | Incremental packing |
| --- | --- | --- |
| Read errors /256 | 112 | 0 |
| Write errors /128 | 83 | 0 |
| Stale journal rejections | 0 | 0 |
| Maximum read ms | 130.132 | 20.136 |
| Maximum write ms | 109.308 | 11.487 |

For experimental packing with visible writes at the high rate, admitted read
p95 is14.822/15.156ms across the two repetitions. This is an observed32-request
quantile per arm, not a population tail estimate or sustainable throughput proof.
Comparisons to the previous artifact are cross-run, not simultaneous pairing.

Without admission the optimized experimental visible-write high-rate arms still
fail:19/32 and20/32 reads fail, with57 and53 stale rejections respectively.
Their p95 values are121.099/134.850ms. All other unadmitted arms pass this finite
screen. Overall30/32 arms pass, not32/32; the full admitted grid passes as
predeclared. There are39 total errors and no cancellation before callback entry.

## Interpretation and remaining scope

The results support a combined rescue: exact computation reduction creates
enough headroom for admission to coordinate writes without excessive queueing
on this workload. Faster computation alone does not solve invalidating-write
contention, and admission alone previously replaced stale errors with deadlines.
Neither observation permits weakening freshness checks or dropping slow work.

This remains an in-memory, finite-burst research runner with a local hash
embedder and repeated public facts. There is no durable-backend confirmation,
randomized arrival phase, long-duration stability proof or full daemon API
installation. Future-compatible writes still wait unnecessarily. All mutations
in the fixture cooperate; unrelated writers/processes remain outside the gate.

Next: carry this exact optimization into a separately frozen durable-backend
experiment and test longer, independently timed arrivals plus mutation coverage.
The seven whole research directions remain open; this closes only the finite
admitted fixed-arrival screen, not goal6 or general agent intelligence.

## Reproduction

INCREMENTAL_SERVICE_PROTOCOL.md preceded both replays. New runners and overlay
sources are hashed in their JSON artifacts. Verifiers check those hashes.

```sh
node research/public-task-pilot/check-incremental-service.mjs
node research/public-task-pilot/check-incremental-openloop.mjs
go test -race -tags researchpriority -overlay research/public-task-pilot/incremental-overlay-v1/overlay.json ./internal/packing ./internal/researchcalendar -count=1 -timeout=120s
```

Raw load data: incremental-openloop-results.json. Public replay data:
each of the five set directories' incremental-service-results.json.
No production or whitepaper changes were made.
