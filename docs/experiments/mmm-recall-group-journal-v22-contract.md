# Guarded SQLite group journal v22: frozen research contract

Date: 2026-10-01. Goal 6 research-only; no production installation.

V21 measured writer-arm publication-guard wait p50/p99 at 11.57/15.68 ms
versus FULL SQLite insert p50/p99 at 0.51/5.56 ms. This experiment tests
whether a single writer permit can validate and durably commit several
independent journal entries without weakening any entry's as-of contract.

## Batch invariants

- Freeze maximum batch size at 8, maximum dwell after the first queued entry
  at 8 ms, and a 500 ms bounded worker guard context. The worker must flush
  pending work on orderly close; no caller receives success before a WAL/FULL
  transaction commits. A canceled caller may receive an uncertain result
  while its queued transaction commits; exact retry must resolve that case.
- The publication store exposes a research-only batch guard which acquires
  its writer permit once, validates **every** `(captured snapshot, as-of)`
  pair against the same current owned publication state, and holds the
  permit across one SQLite transaction. It rejects a batch if any pair is
  invalid. A mixed-horizon negative control must catch validation based
  only on the earliest snapshot or shortest horizon.
- The sidecar coalesces exact same-key/payload retries, rejects changed
  content, retains all journal IDs after close/reopen, and rejects backfill,
  policy motion, unknown history, or missing owner. A transaction error or
  ambiguous commit cannot be treated as success; a later exact retry may
  discover the committed row.
- This remains a single-process test adapter without a durable backend-file
  identity proof. A passing latency screen is not a production claim.

## Load screen

- Compare native LibraVDB journal, single guarded SQLite journal, group
  journal, and group journal with immutable-graph cache. Use the same 200
  as-of-visible records, `RecallK=200`, `PackK=10`, four probe workers,
  192 offers at 8 ms, three paired quiet/writer trials, and 256 future-only
  writes per writer trial. Rotate arm order and reverse quiet/writer order
  on the middle trial.
- Require all offers and writes, exact 200-event nominations, no packed
  future event, 192 journals after reopen, and unchanged graph version in
  the cached arm. Record offer/call/queue and journal-boundary p99, batch
  size distribution, and guard rejections. Any failed entry invalidates its
  arm; do not discard it.
- Predeclared writer offer p99 gate: <100 ms. Even if a group arm passes,
  repeat the original bound-worker/feedback gate, fresh restart, and
  uncertain-commit controls before crediting Goal 6.

Run with `EVENTFRAME_RUN_RECALL_GROUP_V22=1 go test ./internal/service -run
'^TestResearchRecallGroupJournalLoadV22$' -count=1 -v`.
