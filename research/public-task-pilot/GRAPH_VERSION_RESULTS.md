# Persistent graph-state ownership prototype

Status: ownership kernel only, not HNSW, a transaction adapter or performance rescue.

`GraphVersions` stores immutable vertex/vector/edge records in a persistent binary
radix tree over32-bit IDs. Preparing an edit copies33 lookup nodes along its path
and the edited vector/edge payload, sharing untouched subtrees. Up to128 IDs may
be edited per preparation; dimension<=4096 and outgoing edges<=64. The writer
reservation remains held until Commit or Abort. Publication swaps one prepared
root and releases exclusion; no graph search or allocation is present in that
method. No measured allocation benchmark is claimed here.

## Tested invariants

- Prepared changes do not affect current traversal.
- Insertion, changed links and deletion affect only the committed new root.
- Old root traversal retains old vectors/nodes/edges after publication.
- Abort does not change visible state; double settlement fails.
- Input vectors/edges and returned lookup payloads do not alias stored state.
- Editing one half of the key space shares the untouched half's pointer.
- Cancellation and invalid preparation release or preserve writer ownership
  correctly; revision advances only on commit.
- One historical root is traversed1000 times while100 newer roots publish.

The initial targeted tests passed three race-enabled repetitions in1.314s, before
the additional concurrent historical-walk test was added. Final suite below covers
that added test as well.

`go test -race ./internal/researchindex -count=3` passed in21.618s, including
the concurrent historical-walk case. This uses the ordinary researchindex suite.

## Explicit limitations

`Walk` is bounded breadth-first traversal, not nearest-neighbor retrieval.
Deletion removes a vertex but does not repair incoming links; traversal skips
missing targets. Missing targets consume examination budget. A future ANN adapter
must prepare all affected adjacency and entry-point changes or fail its edit
budget; the64-edge limit alone does not bound incoming repair work.

Deletion paths remain allocated until their root/subtree becomes unreachable.
There is no total live-ID limit, old-root lease cap or explicit byte retirement
budget. Go GC handles reachable memory; users retaining arbitrary old roots can
retain arbitrary history. Per-prepare lookup copies are bounded, but total memory
is not thereby constant. Payload cost scales with dimension and degree.

No durable callback is embedded: an external owner must resolve persistence before
Commit and quarantine uncertain outcomes. This is not evidence that a real HNSW
mutation can be staged within the same footprint. Next measure its actual touched
state (including incoming links and global metadata), then test bounded preparation
and memory accounting before a dynamic ANN experiment. All seven goals remain open.
