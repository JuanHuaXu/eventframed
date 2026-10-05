# Handoff Boundary V28: Corrected Prospective Control

2026-10-03. V27 failed its query-belief assertion before writing raw JSON;
its source freeze/test log/run metadata are preserved unchanged. That was an
invalid control placement: changed-query calls occurred in the SAME evidence
epoch, so ordinary member-posterior use could remain valid without any witness
transport. No runtime bug, cross-query transport defect or math correction is
inferred. Original V23/V24/V25 controls applied changed queries AFTER insertion.

New diagnostic order: prime, learned, future-only insert, changed query, changed
vector, changed selection, visible insert. It exercises cross-epoch witness
reuse exactly where intended. Expected beliefs only learned/future-only; other
cases zero. Candidate mechanics, source values, leases, source epochs and
compatibility gates unchanged. V27 control order is not overwritten or replayed
as a successful fresh screen. This technical diagnostic has no quality cohort.

Use the same AST-generated31method EventStore observer and custom journal hook.
Independent AST checker verifies each method observes the correct name FIRST
and delegates to the actual EventStore method; omission of Snapshot is detected.
Do not rely solely on interface embedding/compile success. Dynamic negative
control deliberately calls Snapshot after the boundary and must be observed.

Otherwise V27 protocol unchanged: default config, seven serial isolated native
cases under race, journal full-wire copy/readback, packed-law agreement, no
post-handoff store calls, hashes frozen before run and exclusive outputs.
No early lease release, version-check weakening, power-loss/concurrent mutation
proof, optional asynchronous feature coverage or latency/quality success claim.
Production and all seven original whole-goal definitions untouched.
