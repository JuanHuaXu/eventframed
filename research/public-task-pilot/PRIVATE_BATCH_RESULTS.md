# Bounded atomic insertion batch primitive

`PrivateGraphWriter.InsertBatch` accepts one to four insertion requests, prepares
them serially against private intermediate roots, persists their ID/vector data
and final revision in one callback, and publishes only the final graph. Revision
advances by the number of inserted events; intermediate revisions are never
served. Existing leases retain the pre-batch graph.

Single writes and batches share `persistPrepared`, preserving error/panic
quarantine and durable-success publication. Callback vector slices are owned
copies. All ID-map allocations and retirement capacity are reserved before
persistence. Invalid IDs, duplicates or preparation failure expose no partial
batch and invoke no persistence callback.

## Verification

- Four-event batch produces exactly the same graph records as four serial
  insertions, but calls persistence once and advances revision from 0 to 4.
- Duplicate requests and a bad second vector reject the whole preparation,
  retain revision 0 and leave ID ownership empty.
- Callback error/panic quarantines the owner. Existing single-write lifecycle,
  callback-mutation, late-cancellation and retention tests still pass.
- Abrupt-process recovery now exercises both one- and four-record transactions,
  before and after commit. Three repetitions of each boundary/size yield twelve
  controlled kills: staged batches recover zero inserted records; committed
  batches recover all records, exact vectors and revision 1+batch size.
- Combined race tests passed (2.498s); ordinary package suite passed (5.282s).

## Not yet established

This is a batch primitive, not a request collector. There is no waiting-window
policy, per-request cancellation handling or throughput result yet. The caller
must supply a deadline valid for every batch member. No request may be
acknowledged at enqueue time. Next implement a bounded collector with explicit
oldest-deadline behavior, then repeat the frozen offered-load experiment.

Atomic publication supports safe batching, not a claim that batching meets the
100ms deadline. Long preparation can still consume a batch's latency budget.
The prior failed load and norm-reuse screens remain failed; all seven whole
research directions remain open.
