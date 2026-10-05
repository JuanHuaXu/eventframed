# SQLite journal v20: durable improvement, frozen latency failure

Date: 2026-10-01. [Frozen contract](mmm-recall-sqlite-journal-v20-contract.md).
Apple M4 `arm64`, Go 1.27.1, HEAD `1a7edb6`, dirty research worktree.
The isolated SQLite journal uses WAL/FULL, an as-of commit guard, exact
duplicate/conflict checks, and successful close/reopen counts. Focused
future-only, backfill, policy-motion, owner, and concurrent retry tests pass.
This is still a test-only single-process adapter, not a production storage
design.

| Arm | Quiet offer p99 | Writer offer p99 | Writer call p99 | Writer queue p99 | Writer Search p99 | Writer graph p99 | Writer journal p99 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Native LibraVDB journal | 14.91 ms | 923.74 ms | 70.98 ms | 867.75 ms | 16.79 ms | 21.80 ms | 26.14 ms |
| Guarded SQLite journal | 9.33 ms | **184.89 ms** | 51.29 ms | 145.41 ms | 16.77 ms | 14.72 ms | 23.40 ms |
| Guarded SQLite + graph cache | 9.64 ms | **142.57 ms** | 48.99 ms | 109.72 ms | 16.68 ms | <0.001 ms | 20.06 ms |

Both durable SQLite writer arms **FAIL** the predeclared <100 ms offer p99
screen. Each cell completed 576 exact-as-of Recall offers; writer cells
completed 768 future-only writes. Every SQLite trial reopened with exactly
192 acknowledged journals; no failed commit or guard rejection was removed.
The separate journal removes much of the native backend contention, but the
single synchronous commit and publication guard still leave insufficient
throughput at the 8 ms offer rate. Graph caching helps but does not rescue
the frozen gate. Marginal phase p99s cannot be added to infer a per-request
critical path.

The subsequent [v21 phase trace](mmm-recall-sqlite-phases-v21-results.md)
narrows that explanation: under writers, guard admission p99 is 15.68 ms
while SQLite FULL insert p99 is 5.56 ms. Guard wait, not fsync alone, is the
larger measured journal component on this fixture.

Next test a bounded, acknowledged **group commit** only if one publication
guard can validate each journal's captured snapshot and as-of under the same
mutation exclusion. A batch that validates only the earliest snapshot or
silently acknowledges before fsync would weaken the contract and is not an
acceptable rescue. Also retain a negative backfill/policy-motion control
and an uncertain-commit/restart control. If valid group commit cannot reach
the gate, investigate separate read/write scheduling or a different journal
durability architecture without changing the original success criterion.

Reproduce:

```sh
go test ./internal/service -run '^TestResearchSQLiteJournalContractV20$' -count=1
EVENTFRAME_RUN_RECALL_SQLITE_V20=1 go test ./internal/service -run '^TestResearchRecallSQLiteJournalLoadV20$' -count=1 -v
```

[Raw load log](mmm-recall-sqlite-journal-v20-raw.log) SHA-256
`c615d4ac141f071a1bf20721d00443673519856d883a1d6e01819fae2d7c64da`;
contract `d8982ab6f4167faf542037b2c637117a2bc43dcf4c51e7d01ed8a154243cb198`;
test source at run: sidecar `f524427cc01e4699be82268655ea19b799631260fc3d566734021b7a6fa9ccf9`,
load harness `3e52cbc3fe4cc4d16306f98d4e7fe3b0d5535b2a3e54c3f9aa2902cfd753bffc`.
Goal 6 remains open; no production code changed.
