# Immutable runs with real HNSW graphs

Status: static integration correctness passed; no durable serving or throughput
claim. `internal/researchindex/immutable_run.go` binds each complete copied
manifest (including tombstones) to the private graph built from its live vectors.
`ImmutableRunSearch` captures newest-first run/manifest pairs, precomputes the
shadow-compensated merge plan, and queries each graph with its required prefix.

The adapter uses the existing HNSWBase.Search API and hence caps each prefix at
200, even though the independent merge kernel permits328. Excess shadow demand
fails rather than truncating. Returned IDs are rescored with owned vectors and
merged under latest-run ownership. Recall remains conditional on ANN nomination.

## Tests and observed results

`TestImmutableRunsRealGraphOracle`: three stages, an initial40-record graph,
two overlapping update/delete runs,8-dimensional seeded synthetic vectors and
40 independent query probes per stage. Every returned top10 ID and float64 score
matches exhaustive latest-version cosine ranking. Ingress vectors are overwritten
after construction to test ownership. Historical run subsets remain queryable;
closing a participating run makes subsequent search fail instead of falling back.
The historical-subset check is queryability only, not a separate exhaustive
historical oracle. The stage-specific queries do have exhaustive current oracles.

`TestImmutableRunTombstoneOnlyAndCancellation`: empty-live graph handling, zero
query rejection, cancellation, and closed-empty-run rejection.

Commands:

```sh
go test -race ./internal/researchindex -run TestImmutableRun -count=3
go test -modfile=research-candidate-only.mod -overlay research/public-task-pilot/candidate-only-overlay-v1/overlay-local.json -tags research_candidate_only -race ./internal/researchindex -count=3
```

Both passed (2.771s and16.981s respectively). The first uses the ordinary serial
build/hydrated backend; the second uses the isolated bulk-build/candidate-only
variant and runs the full ordinary plus candidate-contract suite. Each command
includes360 real-graph query/oracle comparisons across three repetitions.
These small graphs do not establish larger-corpus ANN recall or performance.

## Remaining boundary

This object has no authoritative WAL, semantic revision publication, cross-run
leases, concurrent append/drain mechanism, or compaction scheduler. Its caller
must retain all graphs until readers finish. Existing per-graph locking prevents
use-after-close, but a close racing a multi-run query may cause a query error.
Durable publication and bounded retirement must be supplied before serving use.

Next implement run append/publication and compaction with the existing durable
record/revision boundary. Preserve writes arriving during construction and avoid
double-counting drained entries. Bound total runs and retained resources, then
run the unchanged sustained growth load with update/delete controls. No normal
dependency, production config, whitepaper, or frozen experimental output changed.
All seven research goals remain open.
