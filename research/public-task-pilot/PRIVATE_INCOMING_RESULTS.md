# Private incoming-edge discovery

`DiscoverIncomingRemoval` reads immutable target backlinks, identifies live
incoming records that actually reference the target at each level, and produces
owned edits using the backend's swap-with-last removal order. It discovers its
write set from the snapshot; it does not consume a full-graph difference.

This is an explicitly incomplete deletion stage. It does not reconnect neighbors,
retire the target, choose an entry point, persist state, or publish a transaction.
Publishing these edits alone as a complete deletion would be incorrect.

## Evidence

- Three targeted race runs passed in1.341s: read/write-budget rejection returns
  no partial edit list, output mutation cannot change the source snapshot, and
  already-canceled discovery rejects.
- A backend overlay runs removeIncomingConnections at every target level and
  captures states before/after that stage, without later repair/retirement. The
  serial fixture passed in6.474s.
- The private implementation matched every resulting node field across all16
  captured stages at N800/6400. Source links remained unchanged. This explicit
  capture replay passed without race instrumentation in1.730s.

The capture harness intentionally continues from incoming-stage-only states;
it is a stage-equivalence test, not a valid sequence of complete deletions. It
does not validate full ANN recall, crash recovery, or concurrent transactions.

## Bounds and next step

Limits count distinct read ordinals and edited records (up to128). They do not
independently cap duplicate backlink iterations, vector-copy bytes, or elapsed
time. Supplied graph records already have bounded levels/adjacency under the
layered model. Each edited record is copied for ownership; cheaper sharing may
be possible but is not assumed. Incoming edges absent from the recorded backlink
list are outside both this stage and the source backend's accessor procedure.

Next compose bounded private reconnection, target/entry metadata changes, and
one publication owner. The current stage is progress toward real discovery,
not completion of the private update or sustained-load rescue. All seven whole
research goals remain open. Production dependencies are unchanged.
