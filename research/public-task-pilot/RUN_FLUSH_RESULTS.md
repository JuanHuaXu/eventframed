# Captured-version run flush

Status: concurrent flush publication implemented and tested. Consolidation,
bounded historical-reader leases and offered-load validation remain open.

`PrepareRunFlush` captures the pending delta and semantic revision under the
writer gate, reserves the sole build slot, then builds a private immutable run
outside that gate. Writes may proceed during construction. `Publish` reacquires
the gate and carries forward every delta record with revision greater than the
capture. It constructs the matching query plan before changing visible state.
The new graph list, remaining delta and snapshot publish under one gate without
advancing semantic revision or issuing another authoritative write.

Abort, cancellation, failed preflight and quarantined publication close the
unpublished graph. The build slot is released only on successful cleanup. The
prototype permits at most15 graph runs, reserving the virtual delta slot; hitting
that bound requires consolidation rather than another run.

## Tests

`go test -race ./internal/researchindex -run TestRunFlush -count=3`
passed in1.775s.

The full bulk/candidate-only suite also passed three race-enabled repetitions
in17.673s using `research-candidate-only.mod`, the candidate-only local overlay
and tag `research_candidate_only`.

- A blocked real build allows a concurrent committed delete, overwrite and insert.
- A second build is rejected while the first is active.
- Before/after publication rankings match; the newer delta remains intact.
- A second flush drains captured versions without changing revision or results.
- Historical snapshots retain their original answers.
- Double publication fails; cancellation closes the rejected graph and releases
  the slot; quarantine prevents publication and closes its candidate.
- An injected uncertain build failure retains its resource charge.

## Important limitations

The inherited graph builder cannot return detailed failed-close ownership. On a
build error the prototype conservatively keeps its sole build slot occupied.
There is not yet a reconciliation API for that state. This avoids pretending
resources were released, but sacrifices further flush availability. Existing
reads/writes may continue until pending capacity is reached; production readiness
requires a lifecycle owner that can recover or restart this state deliberately.

Published graphs remain in the writer's retained run list. Raw historical
snapshots are not counted leases and must not outlive caller-owned graphs.
This is not bounded-memory reader retirement and cannot yet be used as the
production serving path. Full query-plan preparation still scans manifests.

Next add consolidation publication and bounded reader ownership, then test
multiple consolidation cycles with the unchanged global delta/load budgets.
Static small-run speed evidence does not include these operations. All seven
whole research goals remain open.
