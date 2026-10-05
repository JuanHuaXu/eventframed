# Durable service identity binding v14

RecordedPrediction can now carry an optional immutable ServiceBinding: tenant,
actual service journal/event IDs and dependency snapshot. AdmitBound persists it
with the original experts; retry must match it exactly, including all snapshot
fields. Removing a binding through unbound Admit also rejects. Admission reads
the durable original with detached metadata, so caller mutation cannot alter it.
Replay checks bound tenant identity in addition to numeric learner identity.

The binding test checks changed tenant/journal/event/epoch, removal, caller
mutation, reopen persistence and exact forecast retry. Three full researchmemory
and researchledger race runs plus vet pass, including previous unbound fixtures.
Absent bindings omit the new JSON field, retaining canonical old unbound records.
Older strict readers reject unfamiliar bound records rather than silently drop
their metadata. Binding does not change the prediction math or fitting state.

This is a storage contract, not actual service admission validation. Its test
snapshot is a fixture. The service bridge must verify journal membership,
baseline/features, current dependency compatibility and feedback identity before
using it. The low-level durable worker does not establish provenance, truth or
current snapshot validity. Live service wiring and loaded durable testing remain
open; nothing is configured in the daemon.
