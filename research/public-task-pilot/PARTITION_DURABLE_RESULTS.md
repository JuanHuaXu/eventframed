# Partitioned durable publication reference

Implemented in `internal/researchindex/partition_durable.go`; research-only,
not used by the EventFrame daemon. It retains the existing trusted synchronous
PersistGeneration callback contract and does not supply its own WAL.

## Invariants

- One global semantic revision and one total delta cap across all partitions.
  The64-record cap is not multiplied by the number of partitions.
- Prepare every affected partition and the next global view before calling the
  persistence callback. Invalid input or capacity rejection never calls it.
- The callback must atomically persist all mutations plus the global revision.
  Nil success publishes every prepared core and then the matching global view
  under one admission gate. Callback cancellation after success cannot hide it.
- Errors or panics after callback dispatch quarantine the whole coordinator.
  Aborting unpublished preparations frees exclusion; it does not assert rollback.
  Recovery requires a newly validated authoritative snapshot.
- Old views stay immutable and explicitly historical. Fresh views cannot observe
  half a multi-partition transaction. Partition-local counters are not event
  revisions and must not be compared across partitions.
- At most one compaction builds globally. It copies only the selected partition,
  preserves intervening writes/tombstones and leaves the global revision and
  other partitions unchanged. Publication rechecks quarantine state.

## Evidence

The entire researchindex suite passes three race-enabled repetitions with the
frozen bulk-builder overlay. New tests cover cross-partition blocked publication,
old snapshots, global capacity, invalid second-shard input, lost acknowledgements,
quarantine of prepared compaction, revision overflow, cancellation after success,
callback panic, and200 writes plus200 compactions with1000 concurrent snapshot
checks. All snapshot checks require both members and the global revision to agree.

A real isolated libravdb flat+metadata transaction commits two differently owned
records and the revision, then returns a simulated lost acknowledgement. Fresh
views fail closed; close/reopen finds both exact vectors and the revision, and a
new restored coordinator reads them coherently. This tests that failure branch,
not physical crash recovery or arbitrary transaction-backend correctness.

Command:

```sh
go test -race -overlay research/public-task-pilot/bulk-base-overlay-v1/overlay.json ./internal/researchindex -count=3
```

## Remaining Work

This layer does not publish a matching set of HNSW handles or retire off-heap
resources. An index-serving wrapper must exclude acquisitions while swapping a
compacted partition's view and graph; it must retain old handles for old readers,
charge closing handles against retirement limits, and not expose the cores for
external mutation. The data-only Publish method is not sufficient for ANN serving.

Next implement that wrapper and its lifecycle tests, then rerun the unchanged
growth load, including durable reopen checks. No latency claim is derived from
these tests. Full EventFrame scoring, journals and authority contracts remain
outside this component. All seven whole research goals remain open.
