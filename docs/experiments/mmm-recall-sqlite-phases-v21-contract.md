# SQLite journal phase profile v21: frozen diagnostic

Date: 2026-10-01. Test-only Goal 6 profile; no production edits.

Question: in v20's guarded SQLite journal miss, is the journal boundary
dominated by JSON encoding, pre-insert lookup, waiting for the owned as-of
guard, or the FULL SQLite insert/commit? The guard total contains its wait
and its callback, so its marginal p99 is not additive with insert p99.

Use the same 200-event fixture, `RecallK=200`, `PackK=10`, four probe
workers, 192 offers at 8 ms, three quiet/writer paired trials, and 256
future-only writes per writer trial. Require exact as-of nomination,
no packed future events, all calls/writes complete, and 192 acknowledged
journals after close/reopen. Record offer/call/queue p99 and per-Recall
SQLite encode, lookup, guard-wait, guard-total, and insert p99. A missing
span or guard failure is an invalid run, not a zero-duration sample.

If insert/commit dominates, an acknowledged group-commit experiment is
plausible. If guard-wait dominates, the writer admission architecture must
be addressed first. This diagnostic cannot establish Goal 6 or authorize
weaker synchronous settings.

Run with `EVENTFRAME_RUN_RECALL_SQLITE_PHASES_V21=1 go test ./internal/service
-run '^TestResearchRecallSQLiteJournalPhasesV21$' -count=1 -v`.
