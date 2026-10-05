# Task+lexical full-Recall read-load diagnostic

## Observed result

512 measured requests completed with zero errors. Every returned frontier had
the requested50/200 candidates and every packet's journal matched its session
and explanation. Sixteen arms cover two repetitions, two frontier sizes, one
or four workers, and normal/experimental mode. Arm order reverses in repetition2.

| Frontier | Workers | Control p50 ms, runs1/2 | Experimental p50 ms | Control p95 ms | Experimental p95 ms |
| ---: | ---: | --- | --- | --- | --- |
| 50 | 1 | 4.332 / 4.324 | 4.735 / 4.806 | 5.703 / 4.553 | 4.989 / 4.986 |
| 50 | 4 | 5.723 / 6.028 | 6.181 / 6.224 | 6.098 / 7.082 | 6.835 / 6.734 |
| 200 | 1 | 21.785 / 21.983 | 23.448 / 23.756 | 23.226 / 22.880 | 24.059 / 24.800 |
| 200 | 4 | 25.705 / 25.960 | 28.245 / 28.166 | 27.570 / 28.018 | 30.209 / 28.980 |

Largest experimental duration was30.225ms. This finite read-only sample stayed
under100ms; it is NOT a p99 guarantee or production latency qualification. Each
arm has only32 measured requests. The occasional lower experimental p95 is
sampling/run variability, not evidence that extra work inherently accelerates
the service. Median overhead was roughly0.2-0.5ms at50 and1.7-2.5ms at200.

The200-item arms allocated approximately218-220MB control versus253-255MB
experimental across32 calls. These are cumulative allocated bytes, not retained
heap or resident memory. Allocation/GC cost remains material at sustained load.

## Exact boundary

Go1.27.1, darwin/arm64, GOMAXPROCS4. Hash embedder32 avoids external model/network
latency. Fresh memory store/service per arm; pack10, adaptive and diversity on,
recall equal frontier, budget10000. Two warmups precede the measurement. Unique
sessions force distinct journal records. The duration spans the actual Recall
call, including in-memory journal insertion; journal readback verification is
outside individual timers but inside arm wall time. GC runs before each arm,
not disabled during it.

Four verified public landing facts are repeated under distinct IDs to exercise
cardinality. They are NOT200 independent observations or an accuracy benchmark.
Workers issue requests in a closed loop; there is no externally scheduled arrival
rate or measured upstream queue delay. No ingestion, belief learning, fitting,
graph changes or durable backend writes occur during the measured window.

Accordingly this advances the read-load portion of direction6 only. It does not
close production shadow integration, write-heavy stability, private-agent tests,
or the other six research directions.

## Reproduce

Protocol TASK_LEXICAL_LOAD_PROTOCOL.md preceded dispatch. Raw per-request data,
configuration, allocations and source hashes are in task-lexical-load-results.json.

```sh
node research/public-task-pilot/check-task-lexical-load.mjs
```

The runner cmd/research-task-lexical-load uses the researchpriority build tag and
task-lexical-overlay-v1 Go overlay, then a NEW output JSON path. It never touches
production or prior artifacts. The verifier checks hashes, all512 observations,
frontier/journal assertions and nearest-rank p50/p95/max summaries.

Next: controlled ingestion/version churn and durable persistence under equivalent
load; keep these separate from relation/attribute generalization experiments.
All seven whole goals remain open; no deployment or whitepaper promotion.
