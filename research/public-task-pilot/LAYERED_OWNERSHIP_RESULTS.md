# Layered ownership replay

## Result and scope

`node research/public-task-pilot/replay-layered-ownership.mjs` passed on the
corrected HNSW touch artifact: 32 operations, two corpus sizes (800 and 6400),
and 34 retained snapshots checked against their complete captured node records
and packed global state. The JSON output records input and harness SHA256 hashes.

This is a diagnostic JavaScript persistent-state model, not a Go ANN update
implementation or throughput benchmark. It receives already-computed changes
from before/after snapshots. It cannot establish that discovering/preparing those
changes is cheap or bounded. Vector hashes represent immutable payload handles;
actual float-vector lifetime, allocator state, and backend identifier maps are
outside this experiment.

## Checks

- Layered outgoing links, backlinks, heuristic metadata, IDs, vector hashes,
  levels, and global state are preserved exactly.
- Preparation does not mutate the old root. Publishing a candidate leaves all
  retained roots unchanged, including deleted records in historical snapshots.
- Unchanged records retain pointer identity; changed adjacency with an unchanged
  vector hash retains its immutable payload handle.
- Caller mutation after preparation does not alter the prepared snapshot.
- Edit-budget overflow and adjacency-budget overflow reject unpublished work;
  no links are truncated to satisfy the bounds.

The diagnostic cap is 128 edited records, 32 levels per adjacency array, and
1024 combined outgoing/backlink entries per edited record. Initial construction
uses a separate corpus-sized record allowance. These are experiment limits, not
proved HNSW bounds. Captured operations require 31-113 changed records.

## Cost signal and next experiment

The simple binary radix implementation allocates 33 path nodes per edit:
1023-3729 nodes per sampled update. Copying the complete adjacency arrays of
changed records copies 2810-16274 link entries per operation. This is structural
accounting, not allocated bytes or elapsed-time evidence. Old roots remain
unbounded if retained indefinitely, and deletion does not prune empty paths.

Next, use immutable per-level adjacency lists and actual vector payload ownership
in the Go model, with explicit retained-memory accounting. Then adapt real update
preparation and rerun sustained offered load with durability and reader leases.
Neither full-graph snapshot comparison nor this replay belongs on the hot path.
The previous long-load failures remain unresolved; all seven research goals open.
