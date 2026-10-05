# Composed private insertion

Implemented graph-only `PreparePrivateConnections` and `PreparePrivateInsertion`.
Connections stage forward links/backlinks, reverse diversity decisions and
swap-removal of dropped backlinks on owned records, then prepare an immutable
root. Insertion composes construction search and connection stages from upper
levels down, handles empty/second-node cases and updates the entry summary.

This is serial research code. Caller supplies the sampled level, unique ID and
ordinal, accurate count and coherent summary. Default backend level-zero
selection multiplier is 2.25; physical slack is separate. Scalar cosine is used.
There is no ID-map ownership, durable transaction, bounded reader retention,
asynchronous repair scheduler or concurrent in-flight insertion support here.

## Evidence

- `connect-capture.json`: six independently initialized 80-node backend fixtures,
  levels 0/1, selected-neighbor counts 1/4/8. All resulting records, vectors,
  links, backlinks, heuristic counts and global entry agree; source roots remain
  unchanged. These are explicit selected-neighbor branch controls.
- Existing `serial-work-control.json`: all 16 complete captured insertions at
  N=800/6400 match every output record and global entry. This includes scalar
  private metric calculations, not replayed metric values. Ordinary test passed
  in 1.825s. These are previously exercised fixtures, not fresh confirmation.
- Combined connection and insertion captures, including empty/second-node,
  duplicate ordinal, cancellation, edit-limit failure and immutable-source checks,
  passed with race detection (30.265s).

The first compile found a value/pointer mismatch in owned Lookup handling;
corrected locally before any executable result. No production files were edited.

## Limits and next gate

Per-level edit caps are not a global distinct-edit cap. Evaluation/pair budgets
span the insertion, but do not bound all reads, wall time or retained bytes.
Low-budget failures expose no partial result. ID uniqueness/count/summary
coherence remain caller preconditions, not validated global scans.

Finite graph equality does not prove universal backend equivalence (near ties,
other candidate modes, configuration, concurrent inserts and repair can differ).
Next measure complete preparation cost, integrate insertion with durable private
publication and reader retention, and rerun the original sustained-load failure.
All seven whole research directions remain open.
