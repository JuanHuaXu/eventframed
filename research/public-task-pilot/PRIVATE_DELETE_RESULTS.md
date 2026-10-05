# Composed private graph deletion

`PreparePrivateDeletion` now composes incoming removal and lazy reconnection in
the backend's per-level order, retires the target from the private radix graph,
updates the persistent entry summary, and replaces packed entry metadata when
needed. It returns a private graph/summary pair; it does not mutate or publish
the source snapshot. The distinct edited-record and pair-call caps span all levels.

The incoming-stage helper gained a private one-level mode; its public all-level
operation retains its prior behavior. Reconnection reads accumulated edits from
earlier stages. Target outgoing links supply the initial neighbor list at each
level, then affected incoming ordinals are appended uniquely in source order.

## Verified

All16 complete deletion cases from `serial-work-control.json` at N800/6400 matched
every surviving node field and packed global state. Unlike the previous isolated
reconnection test, this uses actual cosine calculations over owned float vectors,
not a captured metric lookup table. Original source links remain unchanged.
Ordinary capture/singleton/cancellation tests passed in1.901s. The explicit capture
plus private-stage race run passed in26.943s.

The metric is float64-accumulated cosine similarity converted to nonnegative
float32 distance. Matching these fixtures does not prove bitwise equivalence to
all backend SIMD/quantized metric paths, especially near selection ties. Such
paths need declared compatibility tests or the exact backend metric implementation.

## What is still not implemented

- No coherent ID registry, external vector-store ownership, semantic revision,
  durable transaction, or atomic graph/summary publication.
- Summary coherence with the input graph is a caller precondition, not a checked
  full-graph scan during preparation.
- maxReads bounds each incoming discovery stage, not all reads or cumulative
  work. Matrix neighbor limits and pair/edit caps do not imply a total-byte or
  latency bound. Intermediate immutable stage roots allocate before final return.
- Full deletion may reject unsupported/budget-exceeding work rather than silently
  omit repair. A durable caller still needs an explicit fallback protocol.
- Insertion discovery and sustained offered-load testing remain open.

This is the first complete PRIVATE GRAPH deletion in this research path, not a
complete database deletion transaction or daemon feature. Next benchmark combined
discovery/preparation, then wire metadata and publication. All seven whole research
goals remain open; production and normal dependencies are unchanged.
