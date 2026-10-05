# Transaction-scoped write statements v40

Frozen before execution. Isolated ledger experiment, not a serving benchmark.
Twelve rotated cells: three trials, two batch sizes (50 and 200), query-per-record
and transaction-scoped prepared statements. Each fresh ledger executes 32 groups
of admission then typed-discard-shaped terminal writes, in separate FULL commits.
Admission payloads are fixed 1024-byte valid JSON fixtures, not learned forecasts.
Use distinct identities; original bytes and ordered sequences must be checked by
full readback outside timing after every pair. Both modes perform identical work.

Record every admission/terminal duration, source hashes and runtime metadata in
exclusive JSONL. Verify WAL and synchronous=FULL on each ledger. No warmup removal,
retry tuning, excluded trials, production configuration or default API change.
Run prepared/control parity, late-conflict rollback, cancellation, panic,
prepare failure, actual process-exit boundaries and full ledger/learner/service
race/vet checks first. Existing shared cap/identity/JSON validation remains.

Screen: mean pair time must improve by at least 10% at size 200 in every trial
to warrant loaded integration. Report size 50 and p95 as costs/variance, not a
population guarantee. A failed screen is not rescued by selecting favorable
individual operations. SQL prepares may not dominate durable cost. A pass only
supports the next loaded experiment; the 250ms completion-age/serving screens
remain untouched and cannot be established here.
