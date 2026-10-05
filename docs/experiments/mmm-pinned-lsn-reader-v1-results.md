# Pinned-LSN reader v1: Goal 6 component result

Protocol: [mmm-pinned-lsn-reader-v1-protocol.md](mmm-pinned-lsn-reader-v1-protocol.md).
This is an opt-in, test-only LibraVDB reader design screen, not a service or
production change. The comparison is three rotated fresh pairs, each with 200
past-visible records, 256 future-only writes in 16 batches, and 192 concurrent
search offers. Both arms check the captured publication snapshot before and
after each read and verify exactly the 200 base IDs.

| Pair | Order | Current call p50 / p99 | Pinned call p50 / p99 | Current / pinned writer completion |
| --- | --- | --- | --- | --- |
| 0 | current, pinned | 2.497 / 32.878 ms | 0.395 / 28.431 ms | 263.018 / 256.436 ms |
| 1 | pinned, current | 2.662 / 57.106 ms | 0.411 / 27.521 ms | 277.439 / 255.522 ms |
| 2 | current, pinned | 2.699 / 41.349 ms | 0.401 / 28.322 ms | 261.266 / 256.489 ms |

Across 576 reads per arm, current call p99 was **40.153 ms** and pinned call
p99 **27.521 ms** (ratio **0.685**, below the frozen 1.10 ceiling). Offer-to-
response p99 was 40.154 versus 27.522 ms. Summed writer completion ratio was
**0.958** (below 1.25). All arms completed 192 reads and 256 writes; no
future or duplicate result appeared, and per-version future motion held.
The earlier uninstrumented run of the same final test code also passed, with
38.483 versus 28.549 ms pooled read-call p99 (ratio 0.742). A race-instrumented
run passed without a race report; its timings were excluded from performance
claims. Ordinary package tests and `go vet` passed.

Each arm captured durable LSN 254. The temporal API reported 9,728 retained
bytes while the lease was active; this is its archive metric, **not** process
RSS or a steady-state retention measurement. The lease was closed and its
active count returned to zero. An event then written with availability at the
fixed as-of time made `ResearchSnapshotCompatible` reject the old publication
snapshot, although the pinned SQL query still returned the historical 200.
That negative control is load-bearing: serving the old pin without a new
publication guard would silently omit a visible event.

**Decision:** this passes the frozen *future-only reader-path design screen*.
It does not establish Goal 6. Actual chatbot turns generally introduce
immediately visible events. The next experiment must measure a rolling
durable-LSN pin or equivalent fresh view after each visible commit, including
pin/release and index costs, as-of correctness, writer freshness, and serving
tail latency. A fixed historical pin cannot be promoted as the answer.

Reproduce with:

```sh
EVENTFRAME_RUN_PINNED_LSN_READER_V1=1 go test ./internal/store/libravdbstore -run '^TestResearchPinnedLSNReaderV1$' -count=1 -v -timeout 5m
EVENTFRAME_RUN_PINNED_LSN_READER_V1=1 go test -race ./internal/store/libravdbstore -run '^TestResearchPinnedLSNReaderV1$' -count=1 -v -timeout 5m
go test ./internal/store/libravdbstore -count=1 -timeout 3m
go vet ./internal/store/libravdbstore
```

Source SHA-256 at run: test `c1f239a74352f17588bf4e55e08b6bcf1501cd9f75cf912cc0f5278d07c6a4c8`;
protocol `2dd285714c1006d18c7742a8aa7593eef838e593c7bd3d1fbf015b79458d9509`;
store `198a859b3039cabd199d93d975791ae9fe0bfd9fa6dafa6c4dc230ad2fc21974`;
batch writer `ac4d79fda3b0f95f78b917676520f2d9085773766af6dafa6c4dc230ad2fc21974`.
