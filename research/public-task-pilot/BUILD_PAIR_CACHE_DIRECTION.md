# Build-scoped pair-distance cache: kernel only

Inspection of pinned HNSW neighbors.go shows candidate.ID and selected candidate
IDs available in selectWithSimpleHeuristicValues, immediately before repeated
candidate-versus-selected metric calls. The prior profile attributes substantial
CPU to these heuristics. It does NOT establish a high repeat-pair hit rate.

Implemented an isolated BuildPairCache reference in internal/researchindex. It
has at most65536 directed ID pairs, one deterministic float32 callback, and hit/
miss/entry counters. At capacity it computes uncached; it does not evict or grow.
Computations occur outside the mutex; duplicate concurrent misses are permitted.
Directed keys do not assume symmetric floating-point behavior. Separate build
instances cannot share state. There is no pointer-address identity shortcut.

Three race-enabled repetitions pass tests for exact float32 bit preservation,
directional keys, capacity fallback, zero-cache control,8000 concurrent lookups
and separate-build isolation. These are correctness tests, not speed evidence.

## Integration Preconditions

- A cache instance belongs only to a fresh unpublished graph build. Its node IDs
  must retain their vector and metric meanings until every build worker joins.
- The cache must be discarded before graph publication and cannot survive ID
  reuse, vector updates, rebasing or a metric change.
- Connect only the selected-pair heuristic initially, preserving all selection
  decisions. Do not memoize query-dependent distances under node-only keys.
- Measure actual hit rate and overhead with an uncached control, plus same-graph
  distance checks and fresh recall controls. Extra locking/lookup may erase the
  saved arithmetic. No benefit is assumed from the synthetic unit workload.
- Keep the same query effort, graph quality, partition counts and load gates.

The kernel is not wired into HNSW or serving yet. No upstream/dependency code is
changed by this step, no performance claim is made, and all seven goals remain open.
