# Native batch journal v15: large latency rescue, gate still fails

Date: 2026-10-02. The [frozen protocol](mmm-published-recall-batch-journal-v15-protocol.md)
tested a research-only native LibraVDB transaction containing up to
four journals, with at most 1 ms dwell, one readback per entry and
one publication-marker transition per batch. It retained v10's full
service, 256D corpus, 4 ms write/Recall offers, four Recall workers,
read-to-journal admission and all as-of, durability and top-150
oracle checks. Production was untouched.

| Metric | Normal trial 1 | Normal trial 2 | Frozen gate |
| --- | ---: | ---: | ---: |
| Durable writes / successful Recalls | 128/128 / 128/128 | 128/128 / 128/128 | Both 128/128 |
| Stale and semantic violations | 0 | 0 | 0 |
| Journal batches, sizes 1/2/3/4 | 2 / 1 / 0 / 31 | 2 / 1 / 0 / 31 | Report |
| Batch-gate cost p50 / p99 | 7.331 / 8.443 ms | 7.006 / 8.682 ms | Report |
| Journal queue wait p99 | 1.666 ms | 1.274 ms | Report |
| Reader admission wait p50 / p99 | 10.695 / 15.471 ms | 10.413 / 15.262 ms | Report |
| Write offer-to-ack p99 | 42.763 ms | 43.562 ms | <250 ms |
| Recall call p99 | 26.793 ms | 26.299 ms | <100 ms |
| Recall offer-to-done p99 | **154.715 ms** | **155.366 ms** | **<100 ms** |
| Published-view max age | 25.622 ms | 17.779 ms | <250 ms |

The native batch worker formed 31 full four-entry batches in each
trial; its transaction-and-marker gate cost about 7 ms per batch,
versus v14's roughly 6.5 ms for one journal. This is strong evidence
that amortizing the native commit rescues much of the serial capacity.
Offer p99 fell from the v10/v14 roughly 0.5 s range to roughly
0.155 s, but both trials **fail** the unchanged 0.1 s gate. No
threshold was adjusted. The remaining offered-response tail cannot
be attributed to journal queue wait alone: four Recall workers
spent about 10 ms median waiting for read admission, and median
occupied call time was about 21 ms, leaving nominal four-worker
capacity below 250 offers/s. That is a fixture-specific capacity
inference, not a proof that worker count alone will pass.

Focused batch-gate controls passed: exact duplicate and conflicting
same-ID handling, mixed-horizon whole-batch rejection, atomic rollback
after an injected staged insertion, and fail-closed recovery after
interruptions following the LibraVDB and SQLite commits. Every
normal loaded packet and durable journal matched its pinned
snapshot and versioned top-150 oracle. Two `-race` correctness-only
trials also passed all 128 writes and Recalls with no reported race
or semantic violation. A subsequent focused Close-drain check passed
normally and under `-race`: four queued journals were present after
reopening. Instrumented race timings are not
performance evidence. Ordinary package tests and `go vet` passed.

Next test a separately frozen eight-Recall-worker variant while
keeping the native batch cap four, 1 ms dwell, publication rules,
write offers and latency thresholds unchanged. A larger worker pool
could absorb admission wait; it could also increase contention or
writer age. Do not retune v15 or claim Goal 6 complete. The fixture
still lacks durable label-to-forecast learning, crash-recovery load,
large corpora and actual agent outcomes; all seven whole goals stay
open.

Reproduce:

```sh
EVENTFRAME_RUN_PUBLISHED_BATCH_JOURNAL_V15=1 go test ./internal/store/libravdbstore -run '^TestResearchPublishedBatchJournalGateV15$' -count=1 -v -timeout 5m
EVENTFRAME_RUN_PUBLISHED_BATCH_JOURNAL_V15=1 go test -race ./internal/store/libravdbstore -run '^TestResearchBatchJournalDrainV15$' -count=1 -v -timeout 5m
EVENTFRAME_RUN_PUBLISHED_RECALL_LOAD_V15=1 go test ./internal/store/libravdbstore -run '^TestResearchPublishedRecallLoadV15$' -count=1 -v -timeout 5m
EVENTFRAME_RUN_PUBLISHED_RECALL_LOAD_V15=1 EVENTFRAME_RACE_CORRECTNESS_ONLY=1 go test -race ./internal/store/libravdbstore -run '^TestResearchPublishedRecallLoadV15$' -count=1 -v -timeout 5m
go test ./internal/store/libravdbstore -count=1 -timeout 5m
go vet ./internal/store/libravdbstore
```

Git HEAD `1a7edb62b6be4031fd01ebeab8071b17303a7815`, Go
`go1.27.1 darwin/arm64`. SHA-256: protocol
`ae7d6f64f800cdee7cd3a3000f24e82e1e8fe13c80c2a46db73f212d12571191`;
batch gate `fcf677165ba42508cdade8a7d773d74a31f975179bfac55c0dac095b0d59e08c`;
load test (with later Close-drain control)
`62a1d1e73a8bfcb0c29a0d881ff2cd98725fc187c330b92918bac105d293fcb3`.
