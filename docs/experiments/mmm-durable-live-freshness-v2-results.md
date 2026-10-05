# Offered-label to live publication v2

**PASS at finite isolated scope.** The [frozen v2 contract](mmm-durable-live-freshness-v2-contract.md)
corrected v1's timestamp origin without changing the arms, label sequence,
writer workload, or finite thresholds. The [v1 result](mmm-durable-live-freshness-v1-results.md)
is retained as a post-return-only measurement.

Each arm ran three fresh persistent-store trials, 64 guarded bound labels per
trial. Before Close, the notification observer witnessed all 192 labels per
arm reach the worker's published-snapshot count with zero failures or pending
work. The writer arm completed 768 future-dated writes, 762 beginning during
an active learning operation. No future event entered an as-of frontier; the
independent per-trial ledger and same-epoch replay audits passed.

| Execution | Arm | Recall p50 | Recall p95 | Recall p99 | Feedback p99 | Offered-to-observed-publication p50 | p95 | p99 | max |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Ordinary | Quiet | 5.47 ms | 6.45 ms | 7.05 ms | 1.10 ms | 0.71 ms | 1.13 ms | 4.11 ms | 10.02 ms |
| Ordinary | Future writer | 31.44 ms | 43.62 ms | 52.51 ms | 12.29 ms | 7.09 ms | 11.02 ms | 12.29 ms | 12.96 ms |
| `-race` | Quiet | 6.39 ms | 7.41 ms | 9.48 ms | 2.11 ms | 1.73 ms | 2.22 ms | 8.48 ms | 11.37 ms |
| `-race` | Future writer | 143.54 ms | 374.06 ms | 405.83 ms | 122.34 ms | 42.21 ms | 113.36 ms | 122.35 ms | 127.18 ms |

The ordinary-build command
`EVENTFRAME_RESEARCH_PERF_GATE=1 go test ./internal/service -run '^TestResearchGuardedDurableLiveFreshnessV2$' -count=1 -v`
passed both frozen writer-arm screens: Recall p99 <100 ms and live age p99
<250 ms. The corresponding focused service and Durable observation tests
passed under `-race` without applying those ordinary-build timing gates.
The older mixed-write v2 test passed after the shared-helper change;
`go vet ./internal/service ./internal/researchmemory` and `git diff --check`
passed. No production serving path changed.

The observer subscribes immediately before guarded feedback and measures
until it sees the published completion count. Scheduler delay can only make
this observation later, so the metric is conservative for that call path;
it is not the internal fitting duration. The worker may publish before
`Feedback` returns. These p99 values are order statistics from 192 calls
across only three fixtures, not population tails. The writer workload is
future-only; visible backfills/deletes, cross-epoch state transfer, crash
recovery, real outcomes, and end-to-end OpenClaw latency remain untested here.
Goal 6 remains open.

Current source SHA-256: service load helper
`3419f1aa7723e5fca0c0858dcba521664f900c71e339675a7ceac2594dac567c`,
service age test `9b0e60d1fe1d287ff359a741b54fad841b4ac68ee082382e55600b94bddf47c0`,
and read-only Durable observation API
`36d1150bb497b929014dd0171c09fba777cf291a2c9760d3f6efa0ba4f71d035`.
