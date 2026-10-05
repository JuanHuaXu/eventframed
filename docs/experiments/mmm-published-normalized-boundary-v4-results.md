# Published-LSN normalized boundary v4: restricted component passes

Date: 2026-10-02. The [frozen protocol](mmm-published-normalized-boundary-v4-protocol.md)
was run in an opt-in, private four-dimensional LibraVDB fixture. No production
code, default runtime, or existing collection was changed.

The test-only boundary normalized finite nonzero incoming write and query
vectors. It rejected empty, wrong-dimension, zero, NaN and infinite vectors
before storage or SQL. A mixed batch with a valid first event and invalid
second event left the first event absent. All three appended durable vectors
were unit length on readback.

At the new certified LSN, SQL returned five as-of-eligible EventFrames from
the same rows used to score them; the future sentinel was absent. The old LSN
returned only the two genesis rows. A close/reopen retained the new LSN and
five-row result. Converted SQL cosine scores matched ordinary `Store.Search`
on the same committed state both before and after reopen:

| Raw stored contrast | Published similarity | Ordinary similarity |
| --- | ---: | ---: |
| Two aligned genesis rows | 1 | 1 |
| `(3,4,0,0)` | 0.600000024 | 0.600000024 |
| `(0,2,0,0)` | 0 | 0 |
| `(-5,0,0,0)` | -1 | -1 |

The query was deliberately nonunit `(2,0,0,0)`. SQL output was in
nonincreasing similarity order; no exact order was required between ties.
The score bound was `1e-5`, and observed differences were zero at the
displayed precision. Focused normal and race runs, the ordinary package
suite, and vet passed.

This **does not** repair existing nonunit records, enforce normalization at
production ingestion/query boundaries, prove large-corpus top-k parity,
establish loaded service latency, or join candidate rows and Bayesian
journals into one coherent full-Recall view. The v2 nonunit SQL failure
remains valid. Goal 6 and all seven whole research goals remain open.

Reproduce:

```sh
EVENTFRAME_RUN_PUBLISHED_NORMALIZED_BOUNDARY_V4=1 go test ./internal/store/libravdbstore -run '^TestResearchPublishedNormalizedBoundaryV4$' -count=1 -v -timeout 5m
EVENTFRAME_RUN_PUBLISHED_NORMALIZED_BOUNDARY_V4=1 go test -race ./internal/store/libravdbstore -run '^TestResearchPublishedNormalizedBoundaryV4$' -count=1 -timeout 5m
go test ./internal/store/libravdbstore -count=1 -timeout 5m
go vet ./internal/store/libravdbstore
```

SHA-256: test `7693546e483c58eed5cff66d218da9db5b3b6e87d00d5b1c033608f91021940a`;
protocol `ddc1c797d71d1ee27d23ee4e489f91d7559dd1612adce3f4895b90cc9a737f4b`.
