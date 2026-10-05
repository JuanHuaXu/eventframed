# Published-LSN top-k v5: bounded candidate transfer passes

Date: 2026-10-02. The [frozen protocol](mmm-published-topk-v5-protocol.md)
passed in a private 4D LibraVDB collection. All writes used the v4
test-only normalization boundary and publication gate. The collection held
200 as-of-eligible events and 17 future events, including 16 deliberately
high-similarity future distractors. No future row was returned.

At the same certified LSN and committed state, SQL and ordinary
`Store.Search` returned exactly the same candidate membership and cosine
similarity (tolerance `1e-5`) for k=10, 50, and 200. All 32 measured calls
per arm and cap succeeded. The SQL EventFrame bodies decoded and passed
identity/availability checks; its converted similarities were
nonincreasing. Tied aligned rows were compared as a set, not forced into
an arbitrary order.

Quiet call times, 32 samples per arm per cap, alternating order after
four warm calls (nearest-rank p99, thus the maximum of 32 samples):

| k | Published SQL p50 | Published SQL p99 | Ordinary Search p50 | Ordinary Search p99 |
| ---: | ---: | ---: | ---: | ---: |
| 10 | 127.125 us | 346.292 us | 391.625 us | 771.625 us |
| 50 | 258.666 us | 502.292 us | 553.167 us | 877.5 us |
| 200 | 727.625 us | 1.099125 ms | 1.653334 ms | 1.896666 ms |

These timings include the exact-LSN snapshot lease, SQL query and body
decode for the published arm. They exclude ingestion, publication,
journal writes, full Recall, network, and offered-load queueing. The
ordinary Search arm is the existing API on the same small fixture.
Therefore the apparent quiet speed advantage is not a production
throughput or sub-100-ms service claim.

Focused normal and race tests, ordinary package tests, and vet passed.
The result extends v4's five-result score component to finite bounded
candidate sets, but does not establish large-corpus or high-dimensional
ANN parity, loaded tail latency, old nonunit-vector migration, or
candidate/journal read-view coherence. Goal 6 and all seven whole goals
remain open; production was untouched.

Reproduce:

```sh
EVENTFRAME_RUN_PUBLISHED_TOPK_V5=1 go test ./internal/store/libravdbstore -run '^TestResearchPublishedTopkV5$' -count=1 -v -timeout 5m
EVENTFRAME_RUN_PUBLISHED_TOPK_V5=1 go test -race ./internal/store/libravdbstore -run '^TestResearchPublishedTopkV5$' -count=1 -timeout 5m
go test ./internal/store/libravdbstore -count=1 -timeout 5m
go vet ./internal/store/libravdbstore
```

SHA-256: test `fc1ff81b4523681df51cb56a60d6b79e375d380896bc6c8d56a46746a876f014`;
protocol `f945571f0d92e34ca8def6543eeaba84a507eafaef5cf8da14bcf853f86a524f`.
