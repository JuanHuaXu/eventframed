# Original forecast record replay v1

RecordedPrediction captures contract, seed, prediction identity/features/epoch,
time, original inner/outer experts and readiness. Shape/range validation rejects
invalid probabilities, unknown contract, wrong seed/epoch and out-of-order IDs.
RestorePrediction injects recorded experts on a replay-owned Adapter instead of
calling Predict with weights learned afterward. This trusts owned journal data;
shape validity does not authenticate the record or establish its causal timing.

An actual SQLite close/reopen test writes320 admissions and320 feedback records,
then replays pages of37 into a fresh learner. Admissions occur in16-prediction
batches before their outcomes. After320 labels, all512 feature-state forecasts
are bit-identical to the uninterrupted control, and pending count is zero.
This crosses256 lifetime records while respecting the256 concurrent-pending cap.
It does not remove the current service bridge's separate256-journal limit.

Three race repetitions passed for researchmemory and researchledger. Vet first
flagged two unkeyed external Key literals in the test; those were made explicit
and vet passed. Corrupt contract/seed/epoch/order/probability/time, cold-forecast
inconsistency and duplicate restoration are rejected by dedicated tests.

Limits: this is synchronous replay with a trusted typed fixture, not a durable
background admission/acknowledgment pipeline. There is no atomic checkpoint,
crash injection, source fingerprint verification beyond the explicit version,
discard replay or service dependency-authority restoration. Full ordered replay
cost grows with log length. Background snapshots can lag feedback acceptance;
that timing still needs preservation and comparison in the durable consumer.
No production policy or deployment changed.
