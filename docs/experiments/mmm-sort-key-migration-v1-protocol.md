# Sort-key migration v1: frozen private-store feasibility probe

The [rolling-LSN v3 screen](mmm-rolling-lsn-reader-v3-results.md) used a new
private collection with every row born with `available_at_sort`. This probe
tests an existing **private** event collection made by the ordinary
research writer, which has no declared sort-key column and no keyed rows.
It does not touch production data, invoke a production migration, or
publish a serving switch.

Insert three legacy EventFrames available at `.100Z`, `.120Z`, `.125Z`
with distinct IDs and vectors. Save their payload bytes, vectors, runtime
snapshot and per-version motion. First verify the new SQL predicate cannot
bind. Use LibraVDB SQL `ALTER TABLE ... ADD COLUMN available_at_sort TEXT`.
Before backfill, a full-scan readiness check must reject the collection.
Backfill only one row with a parameterized SQL `UPDATE` and show readiness
still rejects it. Backfill the remaining rows with keys derived by parsing
the durable EventFrame `AvailableAt`, not by trusting the old variable-width
metadata string. Readiness may pass only if every row has the exact derived
fixed-width UTC key and correct identity; no inference from a plausible
query result is enough. Query exact latest LSN at `.120Z`, expecting the
`.100Z` and `.120Z` events, never `.125Z`. Reopen and repeat readiness and
query. Check event payload bytes and vectors against the pre-migration
copies; record database LSN and runtime snapshot changes.
Check whether the existing runtime snapshot-compatibility guard notices
the metadata-only backfill; this was added as a diagnostic after the first
run showed record versions advance while runtime version stays fixed.

Negative control: corrupt one key in the private store; readiness must
reject. A pass establishes only a finite backfill mechanism and scan gate.
It does not establish crash-safe publication, concurrent writes, bounded
migration duration, sidecar consistency or full Goal 6. If `ALTER`/`UPDATE`
is unsupported or mutates the event/vector contract, retain the failure
and pursue copy-into-new-collection migration instead.
