# Instrumented HNSW work probe

The separate `work-probe-overlay` instruments query-distance wrapper calls,
published-pair distance calls, link/backlink accessor calls, and entry-point scan
iterations. The pinned backend files and production dependencies are unchanged.
The generator asserts unique patch anchors and records original source hashes.
Counter collection starts immediately before each operation and ends before
after-state capture; initialization and capture reads are excluded.

The 32-operation test passed in 2.491s. Instrumentation uses synchronization and
maps, so that elapsed time is not a usable update-latency benchmark. Full output
is `hnsw-work-probe-results.json`, including operation-local before/after states.

## Results and coverage

- Insertions: 21-162 calls through the instrumented query-distance wrapper.
  This is NOT total distance work: direct/batched/heuristic paths are not all
  intercepted by this wrapper.
- Insertions: up to 470 unique nodes read through the outgoing-link accessor
  and up to 500 through the backlink accessor in a single operation. Sets may
  overlap; do not add their cardinalities to claim a unique total.
- Deletions: 741-2304 published-pair calls. All 16 counts match D(D-1)/2 summed
  over the neighbor sets reconstructed from that operation's own before state.
  `pairRight` records the other endpoint of the SAME call, not another distance.
- No entry-point scan was triggered in this new trace. The previous static audit
  and earlier global-change capture still motivate a forced entry-point case;
  this probe does not dynamically validate the scan bound.

These results reinforce the distinction between a small changed-record set and
the larger search/repair read set. Direct node-array reads, allocation, registry
mutation and other distance paths remain outside the counter coverage.

## Cross-run control and diagnostic failure

A full deep-assert comparison against the old capture stalled without returning
a useful result; that diagnostic process was terminated (exit143), not treated
as a pass. A bounded per-record comparison found extensive differences in the
initial graphs. A fresh UNINSTRUMENTED run also differs from the older control
in 799/800 and 6400/6400 initial records. That control passed in 2.546s and is
preserved in `hnsw-work-control-results.json`.

Therefore exact cross-run topology equivalence is not established, and differences
cannot be assigned to instrumentation from this comparison. A fixed random seed
alone has not produced identical snapshots in these runs. Diagnose initialization
ordering or restore an identical starting snapshot before attempting a controlled
algorithm/performance comparison. The operation-local neighbor/count check remains
valid without claiming identical graphs across runs.

## Next step

Force entry-point deletion, cover all distance/read paths needed for the private
algorithm, and establish an identical-state control. Then implement private update
discovery with explicit read/write budgets; keep slow-path fallback observable.
No sustained-load rescue or whole-goal completion follows from these counts.
