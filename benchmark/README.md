# EventFrame benchmarks

The Go benchmarks exercise production service and storage paths with deterministic
synthetic events. Fixture construction occurs before the benchmark timer. Each
case reports the standard mean plus in-process p50/p95/p99 samples measured with
Go's monotonic clock.

Run the complete local matrix:

```sh
go test ./benchmark -run '^$' -bench . -benchmem -benchtime=1s -count=5
```

Include the expensive 10,000-record durable recall case and the 1,000-record
agency queue ceiling explicitly:

```sh
EVENTFRAME_BENCH_LARGE=1 go test ./benchmark -run '^$' -bench . -benchmem -benchtime=1s -count=1
```

Large fixtures are opt-in because constructing them through the production
transactional ingestion path is intentionally outside the benchmark timer and
can take minutes. Do not confuse that wall-clock index-build cost with the
reported steady-state operation latency.

Run one boundary in isolation:

```sh
go test ./benchmark -run '^$' -bench '^BenchmarkRecall/' -benchmem -benchtime=2s -count=5
go test ./benchmark -run '^$' -bench '^BenchmarkAvailabilityExpansion/' -benchmem -benchtime=2s -count=5
go test ./benchmark -run '^$' -bench '^BenchmarkAgencyClaim/' -benchmem -benchtime=2s -count=5
```

Run the gated concurrent service benchmark through a live LibraVDB contract
sidecar and the durable rank-delta SQLite store:

```sh
EVENTFRAMED_TEST_LIBRAVDB_ENDPOINT=unix:/path/to/libravdb.sock \
  go test ./benchmark -run '^$' -bench '^BenchmarkConcurrentContractRecall$' \
  -benchmem -benchtime=5s -count=3
```

The recall cases include the complete local service path: query embedding,
LibraVDB or in-memory search, Bayesian frontier evaluation, certificate misses,
residual-cache misses, fresh durable decision journaling, reranking, and packing.
The benchmark advances `as_of` by one nanosecond per operation so timed calls do
not collapse into the idempotent-journal retry path. They do not include an
external embedding network request or OpenClaw transport. The gated concurrent
contract case includes `SearchTextCollections`, `RankCandidates`, the embedded
durable event store, and hot persisted rank-delta lookups, but still excludes
HTTP and OpenClaw transport.

The availability cases vary the fraction of records unavailable at `as_of` and
therefore expose LibraVDB's bounded probe expansion. The agency cases hold the
active projection at 1, 100, and 1,000 records and repeatedly claim ten records;
the zero-duration test lease is fixture machinery, not a production setting.

Record the Git commit, Go version, operating system, CPU, memory, power mode, and
whether the run was warm or cold with every published result. Go's benchmark
percentiles and allocation counts are sequential in-process evidence, not
concurrent daemon service-level latency. Production tail claims still require
the separate sustained-load harness planned in the production evidence roadmap.
