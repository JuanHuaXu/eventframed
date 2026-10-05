# Batch open-loop read phases v1: capture dominates

Date: 2026-10-02. Frozen [diagnostic protocol](mmm-sort-batch-read-phases-v1-protocol.md).
Two rotated fresh pairs repeated the same synthetic load with an instrumented
equivalent of the private exact-LSN search. All 256 acknowledgements and
192 reads per arm completed per pair; no integrity violation was reported.
The diagnostic passed its phase-accounting check.

| Pair | Arm | Full read p99 | Publication capture p99 | SQL p99 | Lease p99 | Residual p99 |
| ---: | :--- | :--- | :--- | :--- | :--- | :--- |
| 0 | singles | 12.642 ms | 12.559 ms | 0.541 ms | 0.0006 ms | 0.0056 ms |
| 0 | batch | 15.031 ms | 14.896 ms | 0.634 ms | 0.0015 ms | 0.0039 ms |
| 1 | batch | 14.978 ms | 14.845 ms | 0.623 ms | 0.0010 ms | 0.0039 ms |
| 1 | singles | 11.515 ms | 11.413 ms | 0.458 ms | 0.0005 ms | 0.0578 ms |

The p99 of each phase is not additive to the full-call p99 because the
extreme samples need not be the same call. Nevertheless capture almost
equals full-call latency in both arms, while pinned SQL is sub-millisecond.
The test-only gate holds its lock through database and sidecar publication;
readers calling `capture` wait behind that writer. This is the strongest
local explanation for the failed batch read-tail ratio, not a proof that
all LibraVDB read paths have that bottleneck. It motivates the separate
[last-published-LSN screen](mmm-sort-published-view-v1-results.md).

Reproduce:

```sh
EVENTFRAME_RUN_SORT_BATCH_READ_PHASES_V1=1 go test ./internal/store/libravdbstore -run '^TestResearchSortBatchReadPhasesV1$' -count=1 -v -timeout 5m
```

Protocol SHA-256:
`4f18b40354e0588c1170461e4be8e5443725cbad8a9e0e38a6844c716e91853f`.
Production is unchanged; Goal 6 and all seven whole goals remain open.
