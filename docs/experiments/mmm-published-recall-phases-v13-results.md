# Full Recall phase timing v13: journal lane limits throughput

Date: 2026-10-02. The [frozen diagnostic](mmm-published-recall-phases-v13-protocol.md)
restored v10's full residual-enabled service and added test-only,
per-Recall timing to the unchanged 256D, 4 ms-offer workload. Both
normal trials completed 128/128 writes and 128/128 Recalls with zero
stale, oracle, journal, future-leak or freshness violations. They
still **fail** the unchanged <100 ms offer-to-done p99 gate at
545.249 and 538.402 ms. The instrumented occupied-call p50 was
27.504 and 27.044 ms. Instrumentation changed timing slightly, so
these are diagnostic spans, not an uninstrumented performance claim.

| Per-Recall phase | Normal trial 1 p50 / p99 | Normal trial 2 p50 / p99 |
| --- | ---: | ---: |
| Published-LSN Search | 1.094 / 1.943 ms | 1.101 / 4.832 ms |
| Predictive graph read | 0.001 / 4.187 ms | 0.001 / 4.278 ms |
| Selection + omitted certificates | 0.001 / 4.128 ms | 0.001 / 4.071 ms |
| Residual-read wall interval, 150 calls | 0.349 / 4.753 ms | 0.290 / 4.824 ms |
| Journal owner wait | **12.677 / 21.171 ms** | **12.745 / 20.781 ms** |
| Durable journal append gate | **6.899 / 8.052 ms** | **6.887 / 7.608 ms** |
| Journal view publication | 0.014 / 0.053 ms | 0.013 / 0.016 ms |
| Remaining service work | 1.789 / 2.535 ms | 1.788 / 2.675 ms |

Coverage was 128/128 complete phase records in each trial, with exactly
150 residual-candidate API calls per successful Recall. Residual workers
overlap, so the reported residual interval is first-start to last-end;
their cumulative CPU/blocking time is not added to the call duration.
The remaining-service span subtracts nonoverlapping measured wall
phases and admission wait from each occupied call. Per-phase p50s
must not be summed as if they describe the same particular Recall.

Journal owner wait plus append is the dominant measured median span.
The measured roughly 6.9 ms **serial** append per Recall implies about
145 journal appends/s at that median cost, below the 250 Recalls/s
nominal offer rate even if other work were free. That is a capacity
inference for this single-owner test architecture, not a universal
lower bound for LibraVDB or group commit. Contention from event writes
also contributes to owner wait. The prior [v22 group-journal study](mmm-recall-group-journal-v22-results.md)
failed loaded latency with 8 ms dwell and four workers, so merely
repeating that schedule is not a justified rescue.

The predeclared race-correctness mode passed two further trials with
128/128 writes and Recalls and no reported data race or semantic
violation. Its instrumented timing is not used as normal-performance
evidence. Ordinary package tests and `go vet` passed.

Next: instrument the 6.9 ms append gate internally (marker read,
LibraVDB insert/readback, LSN checks, SQLite marker commit) and test a
different durable journal/publication architecture under the same
4 ms offer schedule. A candidate must keep per-entry as-of validation,
durable-before-ack semantics and a provable event/metadata LSN
relationship; simply omitting journal work would invalidate the Goal 6
contract. All seven whole goals remain open, production untouched.

Reproduce:

```sh
EVENTFRAME_RUN_PUBLISHED_RECALL_LOAD_V13=1 go test ./internal/store/libravdbstore -run '^TestResearchPublishedRecallLoadV13$' -count=1 -v -timeout 5m
EVENTFRAME_RUN_PUBLISHED_RECALL_LOAD_V13=1 EVENTFRAME_RACE_CORRECTNESS_ONLY=1 go test -race ./internal/store/libravdbstore -run '^TestResearchPublishedRecallLoadV13$' -count=1 -v -timeout 5m
go test ./internal/store/libravdbstore -count=1 -timeout 5m
go vet ./internal/store/libravdbstore
```

Git HEAD `1a7edb62b6be4031fd01ebeab8071b17303a7815`, Go
`go1.27.1 darwin/arm64`. SHA-256: protocol
`d89acd3bc1ba41ad700994948438b8519d978c0ee912b26783b5ebcc58335504`;
test `d15f948a95c1ca616f355162f3946ec8344f4457dbd2c1e05f4f7f4f6171e111`.
