# Batch publication open-loop v1: writer rescue, reader gate fails

Date: 2026-10-02. Frozen [protocol](mmm-sort-batch-openloop-v1-protocol.md).
Two unchanged fresh three-pair runs of the private backend screen **FAIL**
the predeclared search p99 ratio limit. Both runs completed all 256 write
offers/acknowledgements and 192 searches per arm in each pair. Actual
median offer gaps were about 4 ms. No search error, future-sentinel leak,
duplicate/unknown ID, lost row, bad sort key or reopen-journal mismatch was
reported. The command exited nonzero in both runs by design.

| Fresh run | Single writer offer-to-ack p99 | Batched writer offer-to-ack p99 | Single search-call p99 | Batch search-call p99 | Read ratio | Integrity violations |
| ---: | :--- | :--- | :--- | :--- | ---: | ---: |
| 1 | 2.261 s | 33.016 ms | 11.776 ms | 15.070 ms | **1.280** | 0 / 0 |
| 2 | 2.186 s | 33.702 ms | 11.429 ms | 14.520 ms | **1.270** | 0 / 0 |

The candidate meets its <250 ms writer-age and <100 ms absolute read-call
limits, but fails the <=1.10 read-tail ratio in both runs. The single-event
control develops a large queue under 4 ms offers; this makes its reads
less exposed to realized commits during the read window. The comparison is
fair at the *offered* schedule and reports the resulting different commit
intensity, but it is not an equal-realized-write-rate isolation experiment.
Do not loosen the ratio after seeing this result. The follow-up
[phase profile](mmm-sort-batch-read-phases-v1-results.md) and
[published-view rescue](mmm-sort-published-view-v1-results.md) are separate
studies, not retroactive passes of this arm.

Reproduce:

```sh
EVENTFRAME_RUN_SORT_BATCH_OPENLOOP_V1=1 go test ./internal/store/libravdbstore -run '^TestResearchIncrementalSortBatchOpenLoopV1$' -count=1 -v -timeout 5m
```

The protocol SHA-256 is
`a6b5234f9af23818bb268bd9825f2721323bb4485054757b4d3d4b10c4a6d059`.
The implementation remains research-only and production is unchanged.
No full Service Recall or background learner was measured. Goal 6 and all
seven whole goals remain open.
