# Durable private graph coordinator

Follow-up: [abrupt insertion and lifecycle checks](PRIVATE_GRAPH_CRASH_LIFECYCLE_RESULTS.md)
now pass at the documented boundaries; the original remaining-gates section
below describes the state before that follow-up. Byte retention and load remain open.

Implemented `PrivateGraphWriter` separately from the deletion-only control.
It serializes writes, prepares immutable insert/delete roots, persists owned
ID/vector mutations plus a revision through `PersistGeneration`, then publishes
the prepared graph/count/summary bundle. Readers can acquire the old version
while preparation or persistence runs. Existing leases survive replacement.

Startup reconstructs ID ownership, count, ordinal high-water mark and entry
summary from a caller-supplied graph. Writes do not scan the whole corpus.
Lease count and retired-version count are bounded. A possible retirement slot
is reserved by rejecting at saturation before preparation/persistence; writer
serialization and preallocated retirement storage prevent post-commit capacity
failure. This is deliberately conservative even if the current root has no
readers at the initial check.

All allocations required for publication precede the callback. Callback mutation
of its vector slice cannot alter the prepared graph. Nil callback return means
durable success, including after late cancellation. Error or panic quarantines
new reads/writes; historical leases remain usable. Preparation errors do not
quarantine. ID-map insertion reserves allocation before the callback; uncertain
state is never reused because the writer then requires reconstruction.

## Verification

- Three race-enabled repetitions of callback tests passed (1.391s): late cancel,
  callback-owned vector mutation, error/panic quarantine, old-version visibility,
  duplicate ID rejection, invalid-vector preflight, idempotent release, reader
  limits and retirement saturation before persistence.
- Real libravdb transaction tests cover insert, a successful insert/insert/delete
  sequence, staged rollback, committed-but-lost acknowledgement and post-commit
  panic. Clean close/reopen verifies durable count, vector and revision. A fresh
  coordinator reconstructs from known fixture IDs. Combined tests repeated three
  times with race detection passed (1.889s).
- Ordinary researchindex package tests passed (5.193s).

## Remaining gates

This is not daemon integration or a production recovery enumerator. The graph
topology itself is not persisted. Known-ID fixture reconstruction is not a full
ANN rebuild. The new coordinator has not yet undergone abrupt-process crash
testing; earlier deletion-only crash results do not transfer automatically.
Lease counts are bounded, but retained bytes, external constructor aliases and
idle lease lifetime are not. Input slices must not be concurrently mutated.
Lifecycle shutdown is not yet implemented. Callback re-entry into write methods
is prohibited because the writer gate is held through persistence.

Next: abrupt-process insertion recovery, shutdown/retention lifecycle checks,
then the original offered-load test with separately recorded Wait/Prepare/Persist
durations. No sustained-load rescue or whole research goal completion is claimed.
