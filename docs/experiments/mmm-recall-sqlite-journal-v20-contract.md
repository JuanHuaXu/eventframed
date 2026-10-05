# SQLite journal sidecar v20: frozen research screen

Date: 2026-10-01. Goal 6, test-only adapter, no production installation.

V18 found that eliminating LibraVDB journal writes removes most queueing;
V19 found that the existing as-of publication guard adds bounded but
nontrivial latency. V20 tests whether a **durable** separate journal commit
retains the gain, without claiming a complete deployable storage design.

## Journal contract

- Use an absolute private SQLite path, one connection, `journal_mode=WAL`,
  `synchronous=FULL`, and a fixed owner token. Creation must not adopt an
  existing file; reopen must reject a missing file or different owner.
- Store JSON journal bytes under `(tenant,id)` with exact duplicate success
  and changed-content conflict. An already committed exact retry may succeed
  after later snapshot motion; a new insert must pass the owned
  `WithResearchAsOfSnapshotWait` check while the SQLite commit occurs.
- A returned successful insert must be readable after close/reopen. Test
  future-only motion acceptance, backfill and policy-motion rejection,
  wrong-owner reopen, duplicate conflict, and concurrent exact retries.
- This test adapter has no cross-process exclusive lock and does not prove
  identity of the LibraVDB backing file from its owner token. Those are
  explicit production blockers, even if this experiment is fast.

## Load screen

- Native durable LibraVDB journal, SQLite sidecar, and SQLite sidecar plus
  immutable-graph cache. For each arm, run three paired quiet/writer trials
  with 200 exact as-of events, `RecallK=200`, `PackK=10`, four probe workers,
  192 offers at 8 ms, and 256 future-only writes per writer trial. Rotate
  arm order by trial and reverse quiet/writer order on the middle trial.
- Require all offers and writes complete, exact nominated IDs, no packed
  future candidate, 192 durable journal rows per trial after close/reopen,
  and unchanged graph version for cached trials. Record offer, call, queue,
  and journal-boundary p99. No failed commit or guard rejection is excluded.
- Predeclared diagnostic latency screen: writer offer p99 <100 ms. Passing
  it is not Goal 6 completion: repeat the original bound-worker/feedback
  gate and crash/restart controls after this no-learner screen.

Run with `EVENTFRAME_RUN_RECALL_SQLITE_V20=1 go test ./internal/service -run
'^TestResearchRecallSQLiteJournalLoadV20$' -count=1 -v`.
