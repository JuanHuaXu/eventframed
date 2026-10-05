# Batch cap v2: smaller batches do not protect readers

The [frozen design screen](mmm-batch-cap-v2-protocol.md) **FAILED** for both
batch-4 and batch-8. Three rotated fresh triples reused the same deterministic
200-past/256-future/192-search workload with the unchanged 4 ms coalescing
bound and the v1 writer/read-tail gates. Every arm accepted 256 unique
writes, completed 192 as-of-correct searches, retained all per-version
motion, and reported no drops or false acknowledgements.

| Pair | Arm | Batches | Ack p99 | Search-call p99 | Writer completion |
| --- | --- | ---: | ---: | ---: | ---: |
| 0 | single | 256 | 958.873 ms | 13.233 ms | 3.212 s |
| 0 | batch-4 | 64 | 308.014 ms | 18.771 ms | 0.920 s |
| 0 | batch-8 | 33 | 178.193 ms | 22.982 ms | 0.499 s |
| 1 | batch-8 | 33 | 163.190 ms | 19.743 ms | 0.485 s |
| 1 | single | 256 | 1023.139 ms | 13.779 ms | 3.384 s |
| 1 | batch-4 | 64 | 302.993 ms | 18.974 ms | 0.898 s |
| 2 | batch-4 | 64 | 306.041 ms | 19.535 ms | 0.904 s |
| 2 | batch-8 | 33 | 198.336 ms | 39.237 ms | 0.522 s |
| 2 | single | 256 | 971.415 ms | 12.996 ms | 3.230 s |

The pooled control acknowledgement/search p99 values were 1015.947/13.195
ms. Batch-4 yielded 305.838/18.974 ms, with writer-completion ratio
0.277, acknowledgement ratio 0.301 and **reader ratio 1.438**. Batch-8
yielded 191.776/23.833 ms, with writer-completion ratio 0.153,
acknowledgement ratio 0.189 and **reader ratio 1.806**. Both writer gates
passed; both reader ratios exceeded the <=1.10 frozen ceiling. In every
fresh triple, each batched arm's search-call p99 was worse than its single
control. No candidate advanced to independent confirmation.

All control arms backpressured nominal 1 ms write offers to about 11 ms
median actual gap. Batch-4/8 median gaps stayed about 1 ms but had p99 gaps
of roughly 13-16/8-9 ms respectively. This repeats the unequal
realized-overlap limitation of v1. The earlier
[factor diagnostic](mmm-batch-reader-factor-v1-results.md) found strong
cross-tenant writer interference under the store's global event lock.
Changing the batch cap alone does not remove that contention. Do not tune
another cap on these consumed timing cohorts; a future candidate needs
different read/write isolation or a formally checked read-aware schedule.

Routine `go test -race ./internal/researchbatch` and `go vet` pass with this
opt-in timing test disabled. The opt-in test exits nonzero by design because
both candidates fail the frozen screen. No production path, full
`Service.Recall`, learner freshness or 4 ms Goal 6 gate was changed or
validated. All seven whole research goals remain open.

Reproduce the expected-failure screen:

```sh
EVENTFRAME_RUN_BATCH_CAP_V2=1 go test ./internal/researchbatch -run '^TestBatchCapV2$' -count=1 -v -timeout 5m
go test -race ./internal/researchbatch -count=1 -timeout 3m
go vet ./internal/researchbatch
```

At-run SHA-256:

```text
9e5fe1daaeed22199ac4dd6b11c8ef27df066223a104d93ebafa2789b853676d  internal/researchbatch/batch_cap_v2_test.go
1670077452e0262dc89a5e8a0e842bf30f27f2ba93a845d1be8930d1541b247d  docs/experiments/mmm-batch-cap-v2-protocol.md
4ce74b342faf48dff70191863466f1b7f47b8ef78f5330e461c0ac6f9fc43dee  internal/researchbatch/batch_ack_load_test.go
```
