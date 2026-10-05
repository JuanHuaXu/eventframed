# Lagging background original-record replay v2

Background.Record now exports the pending original experts under the journal
lock, before feedback consumes that identity. It does not run a new forecast or
wait for fitting. The seed is immutable. Unknown/consumed records reject.

TestReplayPreservesLaggingBackgroundForecasts deterministically holds the fitting
lock while64 predictions/feedback records are issued. Predictions remain on the
cold snapshot even as feedback queues. After release, the worker completes64
labels. Exported records are JSON-roundtripped and restored into a fresh Adapter.

Original-record replay matches the background learner exactly across512 feature
states. A negative control that calls Predict while replaying instead differs
in32 admission expert records and all512 final forecasts. Three repetitions
produced the same counts. This establishes that recomputing historical forecasts
is not a valid substitute for preserving them in this asynchronous fixture.

The targeted race test and three full researchmemory/researchledger race runs
passed; vet passed. This test uses an in-memory ordered operation list with JSON
roundtrips, not a process crash or end-to-end durable background consumer. The
separate v1 test covers SQLite reopening. Their composition still needs direct
integration testing with acknowledgment/crash boundaries and dependency epochs.
No predictive-quality improvement or production readiness is inferred.
