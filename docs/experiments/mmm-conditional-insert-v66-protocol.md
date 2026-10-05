# Conditional-insert admission v66

Frozen before measurement. Fresh admissions use INSERT SELECT WHERE NOT EXISTS
on exact identity/kind, then RowsAffected/LastInsertId. Missing insert means
lookup and exact-byte retry comparison, not permission to ignore a conflict.
Other uniqueness constraints remain errors. Feedback keeps its existing
admission check and path. Same transaction, ordering, FULL WAL and commit-before-
ack; no changes to actual forecast, source identity, guard or replay contracts.
This removes one Go/SQL statement boundary per fresh admission, not all SQLite
index probes. Retry adds a conditional statement and may regress.

Opt-in research owner only; default paths remain controls. Verify receipt and
replay parity, exact/concurrent/within-batch retries, source-index conflicts,
late rollback, feedback ordering, cancellation and panic before measurement.
Run full ledger/memory race tests and vet. Inherited validation/count/byte caps
stay before strategy selection. No dependency or production changes.

Reuse resolved-source isolated workload:50/200 batch sizes, cold/trained64labels,
3 trials,2 rotated strategies,32 fresh/retry/verified-terminal cycles:24 cells.
Freeze per-label barriers and compare complete original/training hashes between
strategies. Record all timings, lifecycle/reopen checks and source hashes in
exclusive JSONL. No competing task-started tests while measuring.

Screen:200-record fresh mean improves>=10% in every paired trial/state; retry
mean must not regress>10% in any paired cell (both sizes). All integrity checks
must pass. Retain failures. A passing isolated screen would still require
scheduled mixed-service confirmation, not default promotion.
