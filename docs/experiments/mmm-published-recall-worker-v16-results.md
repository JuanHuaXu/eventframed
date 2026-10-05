# Eight-worker native batch Recall v16: finite Goal 6 component passes

Date: 2026-10-02. The [frozen protocol](mmm-published-recall-worker-v16-protocol.md)
changed only the research test's Recall worker count from four to
eight. The native batch cap stayed four, dwell stayed 1 ms, and
the same private 256D corpus, 128 visible writes, 128 full Recalls,
nominal 4 ms offers, read-to-journal admission, exact published-LSN
Search, durable journal/readback, as-of gate and top-150 oracle
were retained. Production was untouched.

| Metric | v16 trial 1 | v16 trial 2 | Frozen gate |
| --- | ---: | ---: | ---: |
| Durable writes / successful Recalls | 128/128 / 128/128 | 128/128 / 128/128 | Both 128/128 |
| Stale and semantic violations | 0 | 0 | 0 |
| Journal batches, sizes 1/2/3/4 | 6 / 5 / 4 / 25 | 7 / 1 / 5 / 26 | Report |
| Batch-gate cost p50 / p99 | 6.758 / 9.156 ms | 7.012 / 9.015 ms | Report |
| Journal queue wait p99 | 14.206 ms | 13.731 ms | Report |
| Reader admission wait p50 / p99 | 14.181 / 29.529 ms | 16.054 / 29.068 ms | Report |
| Write offer-to-ack p99 | 66.638 ms | 66.088 ms | <250 ms |
| Recall call p99 | 47.937 ms | 47.026 ms | <100 ms |
| **Recall offer-to-done p99** | **53.153 ms** | **57.731 ms** | **<100 ms** |
| Published-view max age | 25.437 ms | 25.402 ms | <250 ms |

Every normal packet matched its pinned runtime-version top-150
oracle and durable Bayesian journal; no future or
acknowledged-before-offer event appeared. Both normal trials
**pass the frozen finite load gate**. The nearby unchanged v15
four-worker control still failed offer p99 at 170.646 and
151.680 ms, with the same 4 ms offers and 128/128 correct
completion. That comparison supports the hypothesis that worker
capacity, not another journal transaction optimization, was the
remaining bottleneck in this fixture. It does not prove a universal
optimal worker count or a confidence bound for other machines.

Two `-race` correctness-only v16 trials passed counts and semantic
checks with no reported data race. Their much larger instrumented
timings are not normal-performance evidence. The shared v15 native
batch gate passed duplicate/conflict, mixed-horizon, rollback and
interruption/reopen controls. A new focused Close-drain test passed
normally and under `-race`: four queued journals were durably
present after reopening. `go test ./internal/... -count=1 -timeout 5m`
and package `go vet` passed. A repository-wide `go list ./...` is not a
valid clean check because separate research overlay directories
import LibraVDB `internal` packages from outside their allowed parent;
this result does not claim a repo-wide test pass.

This is a **component success**, not Goal 6 completion. The fixture
does not include loaded label-to-forecast learning, graph/posterior
mutation, external writers, large corpora, process crash during the
loaded service, or actual agent outcomes. It uses a test-only
published-LSN adapter and extra worker concurrency on this host.
Those remaining surfaces need separate frozen tests before a
production cutover. All seven whole research goals remain open.

Reproduce:

```sh
EVENTFRAME_RUN_PUBLISHED_RECALL_LOAD_V16=1 go test ./internal/store/libravdbstore -run '^TestResearchPublishedRecallLoadV16$' -count=1 -v -timeout 5m
EVENTFRAME_RUN_PUBLISHED_RECALL_LOAD_V15=1 go test ./internal/store/libravdbstore -run '^TestResearchPublishedRecallLoadV15$' -count=1 -v -timeout 5m
EVENTFRAME_RUN_PUBLISHED_RECALL_LOAD_V16=1 EVENTFRAME_RACE_CORRECTNESS_ONLY=1 go test -race ./internal/store/libravdbstore -run '^TestResearchPublishedRecallLoadV16$' -count=1 -v -timeout 5m
EVENTFRAME_RUN_PUBLISHED_BATCH_JOURNAL_V15=1 go test -race ./internal/store/libravdbstore -run '^TestResearchBatchJournalDrainV15$' -count=1 -v -timeout 5m
go test ./internal/store/libravdbstore -count=1 -timeout 5m
go test ./internal/... -count=1 -timeout 5m
go vet ./internal/store/libravdbstore
```

Git HEAD `1a7edb62b6be4031fd01ebeab8071b17303a7815`, Go
`go1.27.1 darwin/arm64`. SHA-256: protocol
`80a70aa83b6fe78c5e2923e1d6b1beac9336c2b0aa27705d03fa8a1e5daf191b`;
v16 test `b5368316ecf5cd943a2633305c2912aa8915efe4a86400e672d79e5cedfe246b`;
shared v15 batch gate
`fcf677165ba42508cdade8a7d773d74a31f975179bfac55c0dac095b0d59e08c`.
