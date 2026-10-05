# Journal guard cost v19: validation does not recreate native overload

Date: 2026-10-01. [Frozen contract](mmm-recall-guard-cost-v19-contract.md).
Apple M4 `arm64`, Go 1.27.1, HEAD `1a7edb6`, dirty research worktree.
Every cell completed 576 exact-as-of Recall offers; writer cells completed
768 future-only writes. The guarded sink accepted all 192 journals per trial
through `WithResearchAsOfSnapshotWait`; no guard rejection was discarded.

| Arm | Quiet offer p99 | Writer offer p99 | Writer call p99 | Writer queue p99 | Writer journal boundary p99 |
| --- | ---: | ---: | ---: | ---: | ---: |
| Native durable | 14.65 ms | 766.27 ms | 67.13 ms | 714.45 ms | 30.29 ms |
| Unguarded in-memory sink | 4.47 ms | 80.77 ms | 45.03 ms | 42.50 ms | 0.68 ms |
| Guarded in-memory sink | 4.66 ms | 111.56 ms | 45.00 ms | 77.30 ms | 15.12 ms |
| Guarded sink plus graph cache | 4.42 ms | 90.05 ms | 44.05 ms | 57.53 ms | 14.62 ms |

The existing publication guard's writer-arm journal boundary p99 is about
15 ms here, compared with 30 ms for the native LibraVDB journal put. The
guarded sink does not reproduce the native several-hundred-ms queue, but
without graph caching it still misses the <100 ms offer gate. Even the
guarded-plus-cache arm is non-durable and has only a 10 ms numerical margin;
it is a diagnostic, **not** Goal 6 completion. The difference between this
run's unguarded 80.77 ms and v18's 107.62 ms also shows run-to-run load
variability. Do not infer a robust pass from one borderline arm.

Next implement an isolated SQLite/WAL `FULL` journal adapter with exact
duplicate/conflict behavior, as-of guard, and fail-closed reopen. Measure its
commit latency and the original 200-event writer/learner gate; include
backfill, policy-motion, restart, and uncertain-ack retry controls. Do not
alter production journal storage based on this screen.

Reproduce:

```sh
EVENTFRAME_RUN_RECALL_GUARD_V19=1 go test ./internal/service -run '^TestResearchRecallJournalGuardCostV19$' -count=1 -v
```

[Raw log](mmm-recall-guard-cost-v19-raw.log) SHA-256
`e998d61bd9dbe7680882794e7a1a9b6213d8fbbc3fd0fbd0a534c69867f01e42`;
contract `49d8bbde03ee5352411c12302422d8f5dd24484cccc4d627fc123f735f170a81`;
test source at run `8e7637e88a72ec4dc8b7186ff38284f3f6c9c918c2cfdf7e83307bf4468a88c5`.
Goal 6 remains open.
