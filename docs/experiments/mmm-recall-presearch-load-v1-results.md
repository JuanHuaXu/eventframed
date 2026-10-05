# Pre-search snapshot under repeated writes v1: serial service pass

Date: 2026-10-02. Frozen [protocol](mmm-recall-presearch-load-v1-protocol.md).
Two complete fresh-store runs used three rotated blocks per arm and 32
synthetic tenants per block (96 Recalls/arm/run). Every tenant had one
as-of-visible event before Recall. In the write arms, a second event was
inserted synchronously after the first underlying Search and before Recall
read its snapshot. The tested Service code was not changed.

The unmodified-order arm made 96 searches/run and silently returned only
the seed in **all 96 packets/run**, even though each packet and durable
frontier journal claimed the newer Store snapshot. The pre-search pin/write
arm made 192 searches/run: every first attempt was rejected, every retry
returned both eligible events, and packet/journal/current snapshots matched.
The quiet pin arm made 96 searches/run and returned its one event. No
as-of-future event or journal mismatch was observed. This is a repeated
finite consistency result, not a concurrent read-view proof.

All times are end-to-end `Service.Recall` durations in this serial fixture;
the write arms include their synchronous injected Store write. Lower is
better. The unchanged arm is **incorrect**, so it is not a safe latency
baseline. The two runs repeat the same synthetic design and are not an
independent outcome population.

| Run | Arm | Raw Recall p50 / p95 / p99 | Injected write p50 / p99 | Raw minus write p50 / p99 |
| --- | --- | --- | --- | --- |
| 1 | Current/write | 12.098 / 13.925 / 19.240 ms | 6.580 / 13.121 ms | 5.538 / 7.197 ms |
| 1 | Pin/write | 13.400 / 15.563 / 16.911 ms | 6.536 / 8.757 ms | 6.664 / 8.611 ms |
| 1 | Pin/quiet | 5.515 / 6.054 / 6.314 ms | 0 / 0 | 5.515 / 6.314 ms |
| 2 | Current/write | 12.052 / 13.090 / 14.105 ms | 6.488 / 7.621 ms | 5.579 / 6.797 ms |
| 2 | Pin/write | 13.215 / 15.180 / 15.769 ms | 6.508 / 8.313 ms | 6.626 / 9.160 ms |
| 2 | Pin/quiet | 5.514 / 6.151 / 6.931 ms | 0 / 0 | 5.514 / 6.931 ms |

Subtracting write time is only a sequential diagnostic, not an independently
scheduled service latency. The raw p99 arm ordering flipped between runs,
so there is no established latency advantage. The pin/write median incurs
roughly 1 ms more non-write work than current/write in these small stores,
consistent with the extra Search, but this is not a general cost estimate.
The opt-in race run and `go vet ./internal/service` pass; the race run is
not a concurrency stress test because the fixture runs Recalls serially.

This does not connect the certified published-LSN reader or journal-through-
gate transition to full Recall. It also does not validate external
retrieval, graph/certificate/posterior read coherence, loaded p95/p99,
freshness, or learner feedback. Goal 6 and all seven whole goals remain
open; production was untouched.

Reproduce:

```sh
EVENTFRAME_RUN_RECALL_PRESEARCH_LOAD_V1=1 go test ./internal/service -run '^TestResearchRecallPresearchLoadV1$' -count=1 -v -timeout 5m
EVENTFRAME_RUN_RECALL_PRESEARCH_LOAD_V1=1 go test -race ./internal/service -run '^TestResearchRecallPresearchLoadV1$' -count=1 -timeout 5m
go vet ./internal/service
```

SHA-256: test
`5de6aa70d25efaaf53ed22355753d498ce12f08c690d8897723e30797f6fe0cd`;
protocol `4a1b9e1524ef1e2e607da3aa1615e2b3a7246ae96931ead9313e779b3e1f2871`.
