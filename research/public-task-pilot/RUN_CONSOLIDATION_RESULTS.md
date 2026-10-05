# Whole-history run consolidation

Status: consolidation publication correctness implemented; bounded reader
retirement and sustained-load verification remain outstanding.

The consolidation operation captures every run plus the pending delta under
writer exclusion and reserves the same build slot used by flush. It resolves
latest-version ownership off the writer gate, then builds one graph containing
the captured live state. Captured tombstones may be omitted because no older run
is left outside this whole-history consolidation.

Publication rechecks run identity and generation, retains all delta versions
newer than capture, constructs the next query plan, then publishes one replacement
run without changing semantic revision. Writes continue while the graph builds.
Cancellation or quarantine cannot publish a partial layout. Unpublished cleanup
uses the same conservative resource accounting as flush.

## Tests

`go test -race ./internal/researchindex -run TestRunConsolidation -count=3`
passed in1.771s. A controlled real-graph build is blocked while a concurrent
resurrection, deletion and insertion commit. The resulting ranking matches the
prepublication current-state ranking, all newer delta versions remain, and two
captured graphs become one. Captured tombstones do not reappear in the new graph.
Historical rankings remain available on the replaced graphs. Double publication
and a second concurrent build fail. Cancelled publication leaves current state
unchanged and closes its candidate.

The full bulk/candidate-only researchindex suite also passed three race-enabled
repetitions in17.950s using the existing candidate-only local overlay and modfile.

## Ownership and performance limits

Successful publication returns the replaced graph handles without closing them.
This is intentional: raw historical snapshots may still use them. The next
layer must track bounded leases and charge retired resources until actual close
completion. Until then, callers must retain and eventually close graphs themselves.
This primitive alone does not bound historical resource retention.

Whole-history materialization scans/copies retained entries and graph construction
still scales with the consolidated corpus. It only makes that work less frequent;
it cannot establish adequate sustained service capacity. The previous static
screen omitted this cost. A multi-cycle load experiment must include construction,
manifest planning, publication, reader admission and final retirement drain, under
the original delta and deadline limits.

Before-build materialization cancellation releases the build reservation because
no candidate graph has been created. Once the graph builder is entered, uncertain
cleanup keeps the build slot charged, as documented for flush. There is still no
reconciliation API. No production or whitepaper changes. All seven goals open.
