# Go layered-state preparation

## Scope

Research-only `internal/researchindex/layered_graph.go` privately prepares supplied
logical node edits using immutable radix paths, per-level adjacency sharing, and
owned float32 vector slices. Global entry-point state travels with the root.
No daemon code or backend dependency was changed.

The model does not compute HNSW edits, serialize concurrent writers, persist
updates, repair connectivity, or manage reader leases. It does not enforce a
total retained-memory budget. Deletion leaves empty radix paths. These remain
required before a sustained-load rescue can be claimed.

## Replay and ownership evidence

The race-detector run with `RESEARCH_LAYER_CAPTURE` pointing to
`hnsw-touch-results-v2.json` passed `TestLayered*` in 45.543 seconds. All 32 updates
and 34 historical snapshots matched captured logical state. Regenerated actual
768-dimensional vectors matched the capture's float-bit SHA256 hashes.
`layered-go-results.json` records the input hash and per-operation counters.

Three further race runs of `TestLayeredOwnership` passed in 1.319 seconds. Tests
cover caller input/output mutation, private preparation, historical readers,
changed versus unchanged adjacency identity, vector sharing, deletion, global
state, cancellation, duplicate edits, and edit/adjacency capacity rejection.

| Corpus | Copied link entries | Shared link entries | Avoided adjacency copying |
| --- | ---: | ---: | ---: |
| 800 | 48,702 | 66,274 | 57.64% |
| 6400 | 53,858 | 87,576 | 61.92% |

Each corpus's eight insertions copied 6144 vector values total. Other edited
records shared existing vectors. These counters describe the sampled updates,
not a bound for all HNSW mutations. Comparing incoming vectors/levels still costs
time even when their backing storage is shared.

The final ordinary `go test -race ./internal/researchindex -count=1` passed in
8.024 seconds. Capture replay is opt-in and skips in that ordinary invocation;
its explicit run is reported above. `layered-go-source-hashes.json` pins the
implementation, test source, replay output, and input capture.

## Component benchmark

Command (with the same capture environment variable):

```sh
go test ./internal/researchindex -run '^$' -bench '^BenchmarkLayeredCapturedPrepare$' -benchtime=1s -count=3 -cpu=4
```

Apple M4, darwin/arm64, no race instrumentation. Fixed last captured deletion at
N=6400, with 93 changed records. Preparation excludes capture parsing, initial
root construction, edit discovery, verification, durability, and serving.

| Repeat | Iterations | ns/op | B/op | allocs/op |
| --- | ---: | ---: | ---: | ---: |
| 1 | 9284 | 126448 | 118784 | 3578 |
| 2 | 9050 | 127318 | 118784 | 3578 |
| 3 | 9050 | 127887 | 118784 | 3578 |

The first attempted benchmark failed before timing with `research index delta
full`: initialization used the original 6400 cap although the captured state had
6401 records after insertions/deletions. Initial construction now uses its actual
record count; the measured update limit stays 128. This was a harness setup error,
not a throughput rejection. The successful benchmark command exited zero.

## Decision

Per-level sharing is worth retaining for the next research implementation.
The measured component cost is small enough to pursue actual private ANN update
preparation, but is not evidence of sub-100ms end-to-end operation. Reader-lifetime
memory accounting and real update discovery remain the next load-bearing gaps.
All seven whole goals remain open.
