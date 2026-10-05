# Warm and partially applied background recovery v4

The process-recovery fixture now runs cold and warm cases. Warm cases first
complete64 labels, then freeze fitting while recording another64 predictions
and outcomes. Each second-batch record must explicitly report Ready=true; the
first batch must report Ready=false. This avoids a vacuous cold-only test.

Warm-queued exits with64 applied updates and64 durable queued updates.
Warm-applied exits after128 updates. Both recover256 ledger records into128
labels with no pending state and match the corresponding uninterrupted worker
exactly across512 feature states. The two original cold cases also pass.
All four cases passed three race-test repetitions; package vet passed.

This covers mixed applied/unapplied progress at a declared batch boundary and
non-cold original expert records. It does not interrupt inside fitting or commit,
restore a persisted checkpoint, recover a live service bridge, or validate
dependency changes. Full-log replay and trusted fixture identities remain the
mechanism. A live durable consumer and realistic workload validation are still
required; no production change or research-goal completion is claimed.
