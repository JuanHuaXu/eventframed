# Exclusive research ledger v69

Frozen before measurement. v68 identifies shared-memory locking during commit.
Compare resolved-source prepared SQL under normal versus exclusive SQLite
locking. Exclusive mode is opt-in and requires no external readers; research
ownership and single-connection limits already apply. FULL synchronization,
WAL, snapshot guards, original forecasts and receipt semantics stay unchanged.

Pinned modernc.org/sqlite v1.57.0 applies DSN _pragma locking_mode=EXCLUSIVE
before shorthand _journal_mode=WAL and _synchronous=FULL on every connection.
Startup also retains the existing schema/settings setup. Verify locking/journal/
sync readback, forced connection replacement, second-owner refusal, raw reader
exclusion/release and before/after-commit abrupt process exit followed by normal
and exclusive reopen. Process exit is not power-loss simulation.

Same isolated design:50/200 records,cold/trained64labels,3 trials,2 rotated
strategies,32 fresh/retry/verified-terminal cycles:24 cells. Full record/training
hash parity, lifecycle and reopen checks. Capture source hashes and timings
in exclusive JSONL. Run ledger/memory race tests and vet first. No competing
task-started performance work, production access, default/dependency changes.

Screen:200-record fresh mean improves>=10% in every paired trial/state; retry
mean regresses<=10% in every pair at both sizes. Report cleanup too without
substituting it for fresh success. No retuning. Passing isolated cost still
requires scheduled service non-harm and does not finish learning directions.
