# Private bounded reconnection stage

`DiscoverReconnect` now discovers one level's reconnection edits from an immutable
snapshot. It implements lazy matrix creation, the existing min-connection gate,
the source's candidate ordering/selection, reciprocal link/backlink append rules,
physical link capacity including slack, and heuristic reset on outgoing append.
Source records are never modified; returned records are caller-owned copies.

The function bounds neighbor count, edited records, and pair-distance calls.
Budget exhaustion returns no partial edit list. It conservatively rejects duplicate
neighbors and metric errors/nonfinite outputs rather than treating them as a
successfully completed backend update. Those error semantics are intentionally
stricter than the backend's skip-on-distance-error behavior. Callers need explicit
fallback; no silent edge truncation is used to satisfy the edit budget.

## Verification

- Three initial ownership/budget race runs passed in1.344s, including a full
  sparse repair and a dense repeat making zero distance calls.
- Backend fixture captures include exact pair-distance values and full before/
  after graph states for dense, sparse, mixed, single-valid-neighbor, and canceled
  reconnection cases. Capture generation passed in0.224s.
- Three private reconnection plus captured-state race runs passed in1.384s, matching
  all resulting node fields for successful cases and error behavior for cancellation.

Using captured distances isolates graph-update logic from floating-point metric
differences. It does NOT implement or benchmark the production metric callback.
The callback's execution cost is not bounded merely by limiting its call count.
The pair matrix has a separate implicit bound from maxNeighbors (hard maximum1024).
The work limit counts records, not allocated bytes or a deadline guarantee.

## Remaining composition

This is a one-level reconnection stage, not a complete deletion. Incoming removal
and reconnection must execute in the source's per-level order, with accumulated
private edits visible to subsequent work. Target retirement, entry summary/global
state, ID/revision metadata, and durability must publish together. Do not publish
one stage as a whole deletion, or assume all-level incoming removal followed by
all-level repair is proven equivalent without testing.

Next test that composition against complete backend deletes with a real declared
metric. Insertion discovery and the original sustained-load validation remain.
All seven whole research goals are open; production code is unchanged.
