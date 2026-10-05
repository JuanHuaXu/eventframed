# Bounded run-reader lease pool

Status: lease/retirement kernel passed; coordinated publication remains next.

`RunLeasePool` bounds admitted readers and retired graph handles separately.
`Acquire` captures a supplied immutable snapshot and increments references to
every graph it uses. Retiring a graph stops new acquisitions, but admitted
readers retain access. A lease serializes its own search/release operations.
Release clears its snapshot and queues cleanup only when the last reference ends.

Cleanup runs on the existing bounded single-worker retirement queue. Running
and queued cleanup remain charged; successful close removes the retirement
entry. Failed close stays charged and is returned by shutdown. Shutdown refuses
active leases, then stops admission and joins every admitted cleanup. It does
not close the coordinator's current live graphs.

## Evidence

`go test -race ./internal/researchindex -run TestRunLeases -count=3`
passed in1.638s. Tests use real HNSW graphs and a controlled blocking close:

- Reader cap enforced; historical query succeeds after its graph is retired.
- Active leases prevent shutdown.
- Last-reader release returns while close is blocked.
- Retired snapshots cannot be reacquired, even while close is pending.
- A closing graph still consumes the retirement cap.
- Shutdown does not report completion before the blocked close is released.
- Released lease drops graph references and rejects further search/release.
- Injected cleanup failure is surfaced and remains charged.

The full bulk/candidate-only researchindex suite passed three race-enabled
repetitions in18.242s with the existing local overlay and research modfile.

## Integration boundary

This pool accepts a snapshot supplied by a coordinator; it does not decide which
snapshot is current. A coordinator must serialize current-snapshot acquisition
with publication and reserve retirement capacity BEFORE a consolidation publishes.
Calling `Retire` only after publication and ignoring ErrCapacity would be wrong.
That reservation/publication integration is not implemented in this kernel.

The retirement cap counts individual graphs, not whole batches. A consolidation
that replaces9 graphs needs9 retirement slots unless safe reader-free cleanup is
explicitly coordinated. Do not silently reinterpret the earlier load experiment's
resource budget to make a new run layout pass. Any changed graph/memory envelope
must be reported alongside the unchanged pending-delta and deadline gates.

Raw writer snapshots remain available for research tests; a serving coordinator
must not expose them around the lease boundary. External graph close and multiple
independent lease pools owning the same graphs are unsupported. No production
integration or performance rescue is claimed. All seven whole goals remain open.
