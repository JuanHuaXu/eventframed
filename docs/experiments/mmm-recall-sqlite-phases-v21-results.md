# SQLite phase profile v21: guard admission dominates loaded journal tail

Date: 2026-10-01. [Frozen contract](mmm-recall-sqlite-phases-v21-contract.md).
Apple M4 `arm64`, Go 1.27.1, HEAD `1a7edb6`, dirty research worktree.
Three paired quiet/writer trials completed 576 Recall offers per cell,
768 future-only writes in the writer cell, exact 200-event as-of frontiers,
and 192 acknowledged SQLite journal rows per trial after reopen.

| Boundary | Quiet p50 | Quiet p99 | Writer p50 | Writer p99 |
| --- | ---: | ---: | ---: | ---: |
| JSON encode | 0.255 ms | 0.491 ms | 0.236 ms | 0.780 ms |
| Existing-row lookup | 0.254 ms | 0.360 ms | 0.341 ms | 4.285 ms |
| Wait for as-of guard callback | <0.001 ms | 0.002 ms | **11.567 ms** | **15.679 ms** |
| SQLite FULL insert/commit | 0.501 ms | 5.351 ms | 0.512 ms | 5.555 ms |
| Whole guarded callback boundary | 0.502 ms | 5.353 ms | 12.275 ms | 17.869 ms |

The writer arm's full journal boundary p99 is 19.57 ms and offer p99 is
189.91 ms (queue p99 150.42 ms), again failing <100 ms. There is one of
each named span per completed Recall. The callback-boundary row contains
guard wait and SQLite insert; these percentiles are not additive.

This falsifies the narrow hypothesis that FULL SQLite fsync alone dominates
the loaded journal tail on this fixture. Writer-gate admission is the larger
component. A group commit is worth testing only if **each entry** is checked
against its own captured snapshot and as-of while one mutation permit is
held across the transaction. A batch of SQL inserts behind separate guards
would not remove the dominant repeated wait. No journal can be acknowledged
before its durable commit.

Reproduce:

```sh
EVENTFRAME_RUN_RECALL_SQLITE_PHASES_V21=1 go test ./internal/service -run '^TestResearchRecallSQLiteJournalPhasesV21$' -count=1 -v
```

[Raw log](mmm-recall-sqlite-phases-v21-raw.log) SHA-256
`9cc2f45b71f17aeab88e67c0022b5dd28a70d02575965d8cfbfa6195370aee83`;
contract `cc152e1d58e26f96f3f2d5c9e24b1f21d570cb0daf4b1d77588bf2464b75ffa8`;
test source at run: sidecar `d2f05c5c18da53b5840b8df375734120076cfb7afbcae09f3b9da378d3458f84`,
load harness `d5609a986243178214b7702befb3056f8c8b4e14cd6d36bc53b0e4f475e81fd4`.
Goal 6 remains open; production unchanged.
