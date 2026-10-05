# Guarded durable feedback under concurrent future writes v2

**PASS at finite isolated scope.** The [v2 contract](mmm-durable-mixed-load-v2-contract.md)
kept the v1 workload and 100 ms ordinary-build Recall p99 screen unchanged
while separating it from race-instrumented correctness. Test source SHA-256:
`fe247ec3d54f8948b1f8f1533d5efc8b503a08ec83b05b07ca09184c9fe00864`.
The [v1 result](mmm-durable-mixed-load-v1-results.md), including its
instrumented gate failure, remains intact.

Each arm ran three fresh persistent-store trials, each with 64 chronological
guarded admissions and explicit labels. The writer arm completed 768
future-dated writes; 762 began during an active learning operation. All 64
labels per trial were recovered on same-epoch replay with no failed, pending
or queued work. Independent inspection of each trial's 128-row ledger found
unique source journal bindings, expected tenant/event, order and times. No
future event entered the as-of frontier.

| Execution | Arm | Recall p50 | Recall p95 | Recall p99 | Recall max | Feedback p99 | Mean replay of 64 labels |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Ordinary | Quiet | 5.59 ms | 6.31 ms | 7.23 ms | 11.88 ms | 1.76 ms | 8.60 ms |
| Ordinary | Future writer | 32.14 ms | 48.47 ms | 51.97 ms | 52.57 ms | 12.49 ms | 3.55 ms |
| `-race` | Quiet | 6.20 ms | 7.30 ms | 10.61 ms | 13.12 ms | 2.20 ms | 29.81 ms |
| `-race` | Future writer | 143.94 ms | 383.51 ms | 403.16 ms | 404.41 ms | 119.47 ms | 30.38 ms |

The explicit ordinary-build performance command was
`EVENTFRAME_RESEARCH_PERF_GATE=1 go test ./internal/service -run '^TestResearchGuardedDurableMixedWriteLoadV2$' -count=1 -v`;
it passed the finite writer-arm p99 < 100 ms screen. The same test under
`go test -race` passed the correctness and race checks, with the displayed
diagnostic times; `go vet ./internal/service ./internal/researchmemory` and
`git diff --check` passed. No production code or served forecast changed.

The writer arm's ordinary Recall p99 was 7.19x the quiet arm's p99. That
contention is material even though this finite latency screen passed. Each
p99 is an empirical order statistic from 192 calls across only three store
fixtures, not a population bound. The race run is not a comparable
uninstrumented latency estimate. Feedback timing ends at durable submission;
replay after close proves eventual completion, not per-label freshness while
the stream is live. Future-only writes also do not exercise visible
backfills/deletes, epoch transfer, crash recovery, real agent outcomes or
the complete OpenClaw request path. Goal 6 remains open.
