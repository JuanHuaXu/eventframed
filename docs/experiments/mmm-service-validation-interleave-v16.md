# Service validation interleaving v16

The fixture now exercises real Recall/journal/event validation on both memory
and persistent LibraVDB stores. A read-through test adapter first returns the
actual as-of event, then changes the backend policy before returning to the
validator. A no-mutation control must pass and reach the event read; the mutated
case must reach the same read and reject specifically at the final dependency
check. This prevents a vacuous rejection before the intended transition.

Both backends pass three race-test repetitions; service vet passes. No production
code change was needed. This is deterministic interleaving at the GetEvents
boundary, not random concurrent scheduling or proof of linearizable admission.
The validator still returns only point-in-time evidence and does not reserve a
snapshot across subsequent durable writes. Coherent-publication temporal paths,
the whole learner-history cutoff, post-validation mutation and live durable
bridge behavior remain to test/implement.
