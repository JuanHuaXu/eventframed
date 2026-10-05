# Guard phase diagnostic v26

Status: diagnostic supports acquisition amortization, not a completed rescue.
Nine frozen arms repeated v25 with monotonic per-attempt phase measurements.
Three small race repetitions and vet passed before the full run.

| Trial | Mode | Accepted | Prefetch sum (ms) | Entry sum (ms) | Callback sum (ms) | Entry p95 (ms) | Callback p95 (ms) | Queue-age p95 (ms) |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | Single | 74 | 11.727 | 510.818 | 920.692 | 10.789 | 45.076 | 1127.869 |
| 0 | Batch | 185 | 68.054 | 699.274 | 206.711 | 15.940 | 4.320 | 296.486 |
| 1 | Single | 74 | 15.224 | 511.866 | 921.081 | 11.361 | 44.573 | 1124.130 |
| 1 | Batch | 174 | 63.889 | 734.063 | 203.063 | 15.499 | 4.396 | 350.743 |
| 2 | Batch | 190 | 58.105 | 724.876 | 258.487 | 14.670 | 4.714 | 291.632 |
| 2 | Single | 75 | 13.169 | 517.173 | 909.316 | 11.067 | 40.889 | 1118.907 |

Quantiles are nearest rank. Single means the original per-record as-of validator.
Entry includes semaphore wait AND the native snapshot/compatibility check; it
is not a pure lock-contention measure. Prefetch includes request construction.
Queue age is observed at dequeue, not completion. Sums are sequential consumer
durations, not CPU time; quantiles cannot be added to obtain an age quantile.

Batch entry accounts for 73.7-78.3% of total guard time, versus roughly 22-26%
for callback work. Outside-guard prefetch is smaller. All three batch dequeue-age
p95 values already exceed 250ms before admission completes. Thus another small
feature-extraction optimization alone is unlikely to remove the backlog.

Next test one acquisition for a bounded group of already-ready observations,
retaining every member's validation. Include group1 using the same preparation
order as group4, plus original batch and off, to separate grouping from moving
preparation outside the guard. Inspect write latency because longer ownership
can transfer costs. Do not enlarge queues or relax age targets.

All nine raw cells, embedded source hashes, request/write counts, weighted
outcomes and monotonic phase containment were independently checked. Batch
accepted 549/576 with 27 drops; single accepted 223/576, with 352 drops and one
entry expiry. No errors or ledger writes; no labels or fitting.

Artifact: `mmm-batch-phases-v26.jsonl`.
SHA-256: `497817e6785e2884b603be64ffe4429ee7707789cc4ef722d131cc5ff99e3cad`.
