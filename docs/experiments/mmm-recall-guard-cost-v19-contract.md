# Recall journal guard cost v19: frozen diagnostic

Date: 2026-10-01. Research-only Goal 6 diagnostic, no production edits.

V18's in-memory journal sink removed both LibraVDB journal I/O and the
as-of commit validation. A durable sidecar cannot omit the validation. This
screen isolates the cost of the existing owned publication guard before
implementing another database.

## Frozen workload and arms

- Same 200 as-of-visible events, `RecallK=200`, `PackK=10`, four workers,
  192 offers at 8 ms per trial, three trials per arm, quiet and 256
  future-only writer operations. Rotate arm order by trial; reverse the
  quiet/writer order in the middle trial.
- Native persistent journal control, unguarded encoded in-memory journal
  diagnostic, guarded encoded in-memory journal diagnostic, and guarded
  sink plus immutable-graph cache diagnostic.
- The guarded sink calls `WithResearchAsOfSnapshotWait` with the journal's
  captured snapshot and `AsOf` and writes the encoded journal only inside
  its callback. A failed guard must fail Recall, not silently fall back.
- Preserve exact journal conflict behavior in the diagnostic sink. Check
  every journal entry is retained, all offers and writer operations finish,
  graph version stays fixed for cached trials, nominations equal the exact
  live set, and no future packet candidate enters.
- Report offer, call, queue, and journal-boundary p99. Any guard rejection
  is a failed arm, not a discarded slow sample.

## Interpretation

If guarding the sink returns the native overload, moving only the bytes to
SQLite is unlikely to rescue latency under this writer schedule. If guard
cost is bounded, a durable sidecar remains worth implementing. Even a fast
guarded sink is still non-durable and cannot pass Goal 6. No phase-removal
arm is a deployable design.

Run with `EVENTFRAME_RUN_RECALL_GUARD_V19=1 go test ./internal/service -run
'^TestResearchRecallJournalGuardCostV19$' -count=1 -v`.
