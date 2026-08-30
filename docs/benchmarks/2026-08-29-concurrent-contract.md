# Concurrent contract recall benchmark, 2026-08-29

## Result

**Sub-100 ms p95 not validated under this load.** One measured run of the full
Go service path at the benchmark runner's default ten-way parallelism produced:

| Metric | Result | Sub-100 ms status |
| --- | ---: | --- |
| Aggregate benchmark time | 9.064 ms/op | Descriptive throughput only |
| Per-request p50 | 76.91 ms | Validated |
| Per-request p95 | 112.37 ms | Falsified for this run |
| Per-request p99 | 606.13 ms | Falsified for this run |
| Allocations | 68,947/op; 7.97 MB/op | Needs optimization |

The fixture used 200 durable SQ8 events on an Apple M4 with 16 GiB RAM, macOS
26.6.2, and Go 1.27.0. The measured path included embedded LibraVDB event
storage, live LibraVDB 1.9.36-beta.5 `SearchTextCollections` and
`RankCandidates`, Bayesian frontier processing, durable decision journals,
hot SQLite-backed rank deltas, final sorting, and packing. It excluded HTTP,
OpenClaw transport, and network embeddings.

Command:

```sh
EVENTFRAMED_TEST_LIBRAVDB_ENDPOINT=unix:/tmp/libravdb.sock \
  go test ./benchmark -run '^$' -bench '^BenchmarkConcurrentContractRecall$' \
  -benchmem -benchtime=3s -count=1
```

The sidecar had completed the confirmation replay before this run and therefore
was warm, but it was not an otherwise idle freshly restarted daemon. During the
replay and benchmark it repeatedly logged stale micro-temporal dirty-anchor
jobs. On shutdown, LibraVDB panicked with an index-out-of-range error in
`causal.(*DirtyJournal).Drain`. No matching open upstream issue was found on the
test date. This makes the result adequate to reject an unconditional sub-100 ms
p95 claim, not to attribute the tail to EventFrame alone. The next performance
experiment should first use a sidecar build that survives this workload, then
profile queueing, SQLite serialization, journal writes, and background work on
a fresh daemon at declared concurrency levels.
