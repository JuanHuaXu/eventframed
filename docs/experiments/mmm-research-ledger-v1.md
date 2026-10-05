# Durable research ledger v1

New unconfigured internal/researchledger uses the existing SQLite dependency,
one connection, explicit WAL/FULL pragmas and a transactional ordered log. It
does not replace rankdelta storage, the event store or the research bridge.
Use an exclusively owned local research path, not an untrusted/shared path.

Identities include tenant, journal, event and learner contract. Admission and
feedback each have a unique key. Exact byte retries return the prior sequence;
different bytes conflict. Feedback without admission rejects. COMMIT precedes
successful acknowledgment. Bounded256-entry replay pages preserve order without
evicting durable identities. There is no checkpoint or log-deletion API yet.

Three race-test repetitions and vet passed: concurrent identical admission,
conflicting admission/outcome, feedback-before-admission, cross-contract feedback,
ordered pagination, cancelled append and orderly close/reopen with byte-exact
payload preservation. Fixtures contain only explicit synthetic numbers/IDs.

Important limits: payloads are bounded valid JSON, not typed or authenticated
evidence. This layer does not validate original expert forecasts, availability,
dependency snapshots or usefulness. A durable consumer must validate these
before append and on replay. No learner is connected and no restored predictions
have been compared yet. SQL durability settings are not proof of power-loss
recovery: interrupted commits, I/O errors and process termination remain tests
to conduct. Unbounded on-disk growth requires a later tested retention contract.

Next: typed original-forecast records and deterministic reconstruction, followed
by interruption tests and uninterrupted-control comparison. Never recreate an
old prediction using weights learned after its outcome.
