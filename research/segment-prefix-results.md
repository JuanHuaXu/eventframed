# Exact prepared-prefix likelihood prototype

Research-only preparation stores per-start conditional counts, Boolean
agreement counts, log weights and interval evidence for an ordered eligible
prefix. An extension must contain exactly that prefix plus one new sample and
the same family mass. Earlier changes or insertions are rejected, not silently
treated as appends. It is not a completed-result cache: a new outcome is
incorporated for each extension, and every tail forecast is recomputed.

36 complete likelihood/tail table comparisons pass bit-for-bit against full
table-based rebuilding: prefix lengths0/1/16/63, masses0/.95/1, three distinct
appended observations. Changed-prefix rejection checked for nonempty prefixes.
Tests ran with race instrumentation, but concurrent shared-prefix ownership,
cancellation and full snapshot integration still require dedicated tests.

## Component costs

Apple M4, sequential300ms benchmark targets, three repetitions. All samples
are controlled fixtures. Preparation is measured separately, not hidden.

| Operation | ms/op, three runs | Heap B/op | Allocations |
|---|---|---:|---:|
|Full64 likelihood/tails|27.818,27.534,27.885|466944|2|
|Prepared63 + new sample|13.777,13.763,13.798|466944|2|
|Prepare63|27.792,27.970,27.935|6054144-6054145|5|

Each retained interval state is87952 bytes;63 states alone retain5540976 bytes,
plus prefix samples, likelihood matrix and object overhead. Allocation per
extension does not include this retained preparation. Single cold preparation
plus extension is slower than a full fit. Repeated compatible extensions can
amortize preparation, but actual prefix reuse frequency is unmeasured.

This is likelihood construction, not full-fit or service latency. Hazard,
eligible-origin chronology and current as-of still need recomputation by the
full fitter. No support for arbitrary sliding-window evictions, delayed earlier
label insertion or general incremental learning is claimed. Prefix statistics
are private test-only state, not an authenticated cross-tenant cache contract.

```sh
go test ./internal/observationlearners -run '^TestPrefixLikelihoodParity$' -race -count=1 -v
go test ./internal/observationlearners -run '^$' -bench '^BenchmarkPrefixLikelihood$' -benchtime=300ms -count=3
```

Next: ownership/cancellation tests and full posterior integration, then test
service-level changed-prefix controls with preparation cost and memory counted.
Production and previous evidence stay unchanged. All seven directions open.
