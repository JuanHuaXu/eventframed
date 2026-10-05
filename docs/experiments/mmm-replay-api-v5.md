# Validated research replay API v5

ReplayLedger reconstructs one exclusively owned, quiescent research stream.
It checks tenant, stream, contract, decimal learner prediction ID binding,
record order and typed feedback legality. Exact canonical encoding rejects
unknown/omitted fields, duplicate JSON keys and trailing payload data. A failure
returns no partially reconstructed Adapter. This is validation, not provenance
authentication; the caller must independently establish journal ownership and
service dependency authority before using any reconstructed result.

Three researchmemory/researchledger race runs and vet pass. Negative fixtures
cover tenant/stream/contract/ID mismatch, absent usefulness, unknown fields,
early feedback and duplicate JSON keys. Missing usefulness cannot default into
negative evidence. All four cold/warm process-exit recovery cases now also call
the reusable API and match the uninterrupted control across512 feature states.

Research ledger Event keys currently name learner prediction IDs, not actual
service event IDs. A service-facing consumer requires an explicit separate
binding. Replay requires stopped writers, has no checkpoint and scans the full
log. There is no live acknowledgment pipeline, dependency revalidation, discard
operation, retention scheme or restart-cost bound. This is an implementation
step toward durable learning, not a completed continuous-learning system.
