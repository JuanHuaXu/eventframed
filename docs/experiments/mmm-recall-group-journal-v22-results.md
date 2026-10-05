# Guarded group journal v22: valid batching worsens loaded latency

Date: 2026-10-01. [Frozen contract](mmm-recall-group-journal-v22-contract.md).
Apple M4 `arm64`, Go 1.27.1, HEAD `1a7edb6`, dirty research worktree.
The research-only batch guard validates every journal's captured snapshot
and as-of under one mutation permit. Mixed-horizon backfill and policy
controls reject; the WAL/FULL worker acknowledges only after transaction
commit. The full-queue mutex regression and focused race tests pass.

| Arm | Quiet offer p99 | Writer offer p99 | Writer call p99 | Writer queue p99 | Writer journal p99 |
| --- | ---: | ---: | ---: | ---: | ---: |
| Native LibraVDB journal | 14.80 ms | 974.57 ms | 71.29 ms | 917.64 ms | 32.45 ms |
| Single guarded SQLite journal | 12.32 ms | 203.30 ms | 54.48 ms | 164.73 ms | 25.41 ms |
| Group journal | 22.95 ms | **642.17 ms** | 66.51 ms | 593.21 ms | 29.49 ms |
| Group journal plus graph cache | 23.05 ms | **669.83 ms** | 57.34 ms | 616.37 ms | 30.09 ms |

Both group arms **FAIL** the predeclared <100 ms writer offer p99 gate,
and both are worse than the single SQLite sidecar. Every cell completed
576 exact-as-of Recall offers, every writer cell completed 768 future-only
writes, and each trial reopened with 192 acknowledged journal records.
There were zero batch-guard rejections. The group worker did form batches:

| Arm | Quiet batch sizes 1/2/3/4 | Writer batch sizes 1/2/3/4 |
| --- | ---: | ---: |
| Group | 133 / 214 / 5 / 0 | 118 / 96 / 78 / 8 |
| Group plus graph cache | 118 / 223 / 4 / 0 | 13 / 183 / 3 / 47 |

With only four Recall workers and 8 ms offered arrivals, the frozen 8 ms
batch dwell and single batch worker add queueing faster than shared guard
acquisition amortizes it. This is a measured explanation, not proof that
every group-commit schedule must fail. Tuning dwell or worker count on this
same fixture would be a new design experiment, not a retroactive pass.
The next credible work is to separate as-of validation from the long event
writer critical section while preserving a durable total order, or to test
a materially different journal/publication architecture. Do not weaken
FULL durability, per-entry as-of checks, or the <100 ms gate.

Reproduce:

```sh
go test ./internal/researchpublicationstore -run '^TestResearchAsOfBatchGuard' -count=1
go test -race ./internal/service -run '^TestResearchSQLiteGroupJournal' -count=1
EVENTFRAME_RUN_RECALL_GROUP_V22=1 go test ./internal/service -run '^TestResearchRecallGroupJournalLoadV22$' -count=1 -v
```

[Raw load log](mmm-recall-group-journal-v22-raw.log) SHA-256
`180871cac8076879451ef999b635c7ceb3b1c5dadf34dc4c184669de7ed00a91`;
contract `9594dc26bf6d81b3abb501c524e11598330ef07d9ed4bf35c013f825f865fa78`;
test source at run: batch guard `b4ad200efcafba1c97bce12a2c27e4ea49e32989f0608f5f9f0ddfbe120bb168`,
guard tests `859c05637a06d0a498accd08c384ef78ff23db3d4bf525e117ae71bad60a32dd`,
group worker `e77612716407dbc8f476315f3c5d66e6d951e8813de7c20f9f9aeb04dacc545a`,
load harness `f295522f00333b6a5f408e100a38d4e1a271415d9a60e0dc99e0e98236111e81`.
Goal 6 remains open; production unchanged.
