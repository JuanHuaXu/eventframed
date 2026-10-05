# Concurrent pre-search snapshot v1: frozen screen fails

Date: 2026-10-02. Frozen [protocol](mmm-recall-presearch-concurrent-v1-protocol.md).
Two fresh rotated pairs ran actual `Service.Recall` against private LibraVDB
stores. Each arm offered 64 requests per block at a median gap near 4 ms
to four workers, with a separate synthetic tenant, an eligible seed, and
a future-available sentinel per request. One eligible event was committed
immediately after that request's first Search. The pin arm used request-
scoped test state, not the serial wrapper from the earlier study.

| Run | Arm | Success / stale / other | Searches | Silent omissions | Call p50 / p95 / p99 | Offer-to-response p50 / p95 / p99 |
| --- | --- | --- | ---: | ---: | --- | --- |
| 1 | Current | 128 / 0 / 0 | 252 | **44** | 36.009 / 65.964 / 70.858 ms | 223.481 / 405.866 / 424.982 ms |
| 1 | Pin | 125 / 3 / 0 | 341 | **0** | 38.978 / 85.941 / 96.026 ms | 227.759 / 412.316 / 434.554 ms |
| 2 | Current | 128 / 0 / 0 | 247 | **45** | 35.982 / 64.923 / 66.039 ms | 220.237 / 395.044 / 414.928 ms |
| 2 | Pin | 124 / 4 / 0 | 335 | **0** | 38.971 / 83.132 / 105.754 ms | 239.820 / 416.030 / 437.742 ms |

Every offered request committed its injected write exactly once. No
future-available or cross-tenant candidate, journal/packet mismatch, or
other error was observed. An omission means a successful packet claimed a
runtime version at least as new as its own injected event but omitted that
event. The current arm is therefore an incorrect control, not an acceptable
lower-latency design. The pin arm prevents that observed omission but fails
both frozen gates: success is below 127/128 and offer-to-response p99 is
far above 100 ms. This is **not** a Goal 6 pass.

The race-instrumented run reported no data race and failed the same gate:
pin 127/128 successes, zero omissions, offer-to-response p99 503.973 ms.
Its timings are not comparable with the ordinary runs. Ordinary service
package tests and `go vet` pass. The two ordinary runs repeat the same
synthetic design rather than furnishing independent task outcomes.

Interpretation: concurrent as-of-visible writes cause more retries under
the pin. The [cross-tenant probe](mmm-recall-cross-tenant-retry-v1-results.md)
proves that global runtime motion can trigger one such retry even when the
queried tenant's candidate set is unchanged. That is only part of the
story: service-call p99 approaches or exceeds 100 ms in the pin arm, and
offer-to-response p99 exceeds 400 ms in both arms because queued requests
outpace this fixture's completion capacity. Neither a different retry
threshold nor a tenant-scoped guard alone proves a throughput rescue.

The test uses ordinary Store Search and journal operations. It does not
connect the certified published-LSN reader, batch publication, or
journal-through-gate metadata transition. A coherent immutable/as-of read
view, relevant-scope invalidation, and bounded write/admission architecture
remain required. Production was untouched; all seven whole goals are open.

Reproduce the expected gate failure:

```sh
EVENTFRAME_RUN_RECALL_PRESEARCH_CONCURRENT_V1=1 go test ./internal/service -run '^TestResearchRecallPresearchConcurrentV1$' -count=1 -v -timeout 5m
EVENTFRAME_RUN_RECALL_PRESEARCH_CONCURRENT_V1=1 go test -race ./internal/service -run '^TestResearchRecallPresearchConcurrentV1$' -count=1 -v -timeout 5m
go test ./internal/service -count=1 -timeout 5m
go vet ./internal/service
```

SHA-256: test
`680a3f0cc9d9cc7bcfadac79f82ed1e5ecfbf9802ef4e2aabe133114fa7988ad`;
protocol `e76b557956a7e967454532b4dbc45f51e380492cf8132705cd72edf141b475f1`.
