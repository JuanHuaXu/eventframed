# Read-to-journal admission v10: completion rescued, latency not rescued

Date: 2026-10-02. The [frozen protocol](mmm-published-recall-admission-v10-protocol.md)
used the same 256D corpus, 4 ms write/Recall offers, four Recall workers,
cap-16 writer, durable frontier journal and top-150 version oracle as
the failed v9 screen. A test-only Go RWMutex admitted each Recall for
its whole Search-to-journal interval and admitted event batches as
writers. Production was untouched.

| Metric | Normal trial 1 | Normal trial 2 | Frozen gate |
| --- | ---: | ---: | ---: |
| Durable write acknowledgements | 128/128 | 128/128 | 128/128 |
| Successful Recalls | 128/128 | 128/128 | 128/128 |
| Stale rejections / Search calls | 0 / 128 | 0 / 128 | 0 stale |
| Writer admission-wait p99 | 31.410 ms | 32.836 ms | Report |
| Reader admission-wait p99 | 34.096 ms | 33.698 ms | Report |
| Write offer-to-ack p99 | 88.075 ms | 85.384 ms | <250 ms |
| Recall call p99 | 62.104 ms | 62.574 ms | <100 ms |
| Recall offer-to-done p99 | **515.507 ms** | **556.699 ms** | **<100 ms** |
| Published-view maximum age | 31.436 ms | 32.851 ms | <250 ms |

All successful packets and journals matched the pinned snapshot and
runtime-version top-150 oracle; no future or acknowledged-before-offer
violation appeared. The admission lock removed v9's 10/13 exhausted
Recalls without breaching the writer freshness gate. It did not make
the service keep up with 250 Recall offers/s: median occupied Recall
time remained about 26-28 ms, so four workers' nominal capacity is
only about 143-154/s before queueing. The offers continued near 4 ms;
the offer-to-done p99 stayed around half a second. Compared with v9,
the response tail changed little and writer age increased. This is a
capacity inference from the measured fixture, not a proof about all
worker counts or journal architectures.

The predeclared `-race` correctness-only mode passed both full trials:
128/128 writes and Recalls, zero stale/oracle/journal/future violations,
and no reported data race. Its much larger instrumented timing numbers
were reported but not used as normal-performance evidence. Ordinary
package tests and vet passed.

The **overall frozen v10 screen fails** because both normal
offer-to-done p99 values exceed 100 ms. A next test should profile
Search, scoring/packing, journal owner wait, durable journal commit,
and admission wait under the same offers before selecting a new
architecture. Faster isolated SQL in v6/v7 and eliminating retries
here are insufficient by themselves. Goal 6 and all seven whole
goals remain open.

Reproduce:

```sh
EVENTFRAME_RUN_PUBLISHED_RECALL_LOAD_V10=1 go test ./internal/store/libravdbstore -run '^TestResearchPublishedRecallLoadV10$' -count=1 -v -timeout 5m
EVENTFRAME_RUN_PUBLISHED_RECALL_LOAD_V10=1 EVENTFRAME_RACE_CORRECTNESS_ONLY=1 go test -race ./internal/store/libravdbstore -run '^TestResearchPublishedRecallLoadV10$' -count=1 -v -timeout 5m
go test ./internal/store/libravdbstore -count=1 -timeout 5m
go vet ./internal/store/libravdbstore
```

SHA-256: test `f587e947ec08370d8bf9da8f2131e2a14bbc51eefb1949f06d6ac1a570ec812c`;
protocol `33c289965c74aa39ae3c336a6e1941843dd3148a4089999fff95c344849ef9b9`.
