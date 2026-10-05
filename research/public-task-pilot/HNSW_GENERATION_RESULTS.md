# Real immutable HNSW read adapter and guarded compaction

Added durable_compaction.go and hnsw_base.go under internal/researchindex. These
remain unwired research components, not a replacement event store.

## Verified behavior

Coordinator-managed compaction checks health before building and again under
the write gate before publication. A candidate captured before an uncertain
write cannot publish afterward. Canceled publication releases the compaction
slot. Capacity rejection does not reach persistence; compaction allows later
writes while retaining earlier records. Newer updates/tombstones survive rebase.

The ANN adapter builds an actual libravdb HNSW collection in a new private
directory from the immutable base. It queries k+delta_count candidates (bounded
by base size), re-scores returned IDs with the same cosine formula as the delta,
then uses the tested shadowing merge. It checks base-generation identity and
rejects searches after explicit close. Empty-base/delta-only queries work.

Tests demonstrate deletion of a leading base candidate, demotion of another,
promotion of a delta record, old-view stability, generation mismatch rejection,
and successful results after building a replacement HNSW base. Three race-enabled
full-suite repetitions pass: `go test -race ./internal/researchindex -count=3
-timeout=120s`. These tiny controlled examples are correctness tests, not broad
ANN recall validation or performance evidence.

## Remaining serving gap

The test explicitly publishes compaction and then builds the replacement ANN.
During that gap the old adapter rejects the new view. A production path must
build the index before publishing an atomic generation/index pair, or explicitly
reject reads while unavailable. It must never quietly combine a new delta with
an old base. Reader lifetime management must also keep retired graph resources
alive until existing readers finish; a Close mutex only guards active calls.

This builder inserts records sequentially into a separate derived database. Its
build I/O and whole-corpus resource costs have not been optimized or benchmarked.
The authoritative durable coordinator uses a separate correctness fixture; the
combined ANN/store service has not yet been implemented. No general scalability,
latency, semantic recall, crash recovery or production-readiness claim follows.

Next: prebuild the ANN for a captured compaction, publish the generation/index
pair under one admission boundary, preserve intervening delta writes, and test
old-reader retirement before measuring the full pipeline.
All seven whole research goals remain open; production remains unchanged.
