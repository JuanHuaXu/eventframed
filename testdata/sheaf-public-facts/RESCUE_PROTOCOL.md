# Signed-Snap Rescue Protocol

This protocol was frozen before regenerating or inspecting the signed-snap
aggregate results.

## Fixed implementation

- Existing edges with no `effect` retain symmetric `compatible` behavior.
- `supports` propagates the source node's bounded score to the target only.
- `supersedes` propagates one minus the source node's bounded score to the
  target only. It is a predictive rank relation, not a causal edge.
- Rank corrections retain the existing clipping, boundary-plasticity, and
  correction-reliability controls.
- Split publication continues to invalidate the declared reverse dependency
  closure atomically. The bounded nomination frontier is not suppressed.
- The primary run uses graph weight 0.05, recall frontier 16, pack size 3, and
  the deterministic 256-dimensional development hash embedder.
- Graph weight 0.25 is a sensitivity run only and cannot rescue a failed primary
  result.

## Frozen benchmark repairs

- Turn 10 uses a deterministic paraphrase rather than repeating turn 5 exactly.
- Turn 5's prior correct answer is included in turn 10's relevance oracle.
- The correction bucket includes turn 5.
- `harmful_overmerge` has an active wrong directional support edge and must not
  be interpreted as a no-graph control.

## Controls

- `current` and `unchanged` must be identical query by query.
- A unit test must show directional supersession demotes only its target.
- A service integration test must show a negative target rank delta and complete
  rollback.
- No candidate may nominate an event outside the bounded retrieval frontier or
  include a graph member unavailable at publication time.

## Primary outcome rule

Compare `split_and_rewire` with `current` on the eight-session, 16-query
confirmation split. The retrieval rescue passes only when:

1. relevant-hit rate is no lower;
2. obsolete-hit rate is no higher;
3. at least one of those two rates improves strictly; and
4. mean reciprocal rank is no more than 0.05 below `current`.

Failure of any condition leaves the retrieval rescue unvalidated. Design-split
performance and the 0.25 sensitivity run are diagnostic only. This experiment
does not validate snap publication admission, Anti-Pigeon certificate coverage,
Bayesian posterior pooling, a production semantic embedder, or causal edges.
