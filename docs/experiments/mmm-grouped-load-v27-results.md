# Ready-group admission v27

Status: finite age/read-latency screens PASS for group4; full integration remains
unproven. This is an uninstalled research consumer, not a production scheduler.

The v26 timing diagnostic motivated amortizing guard entry across at most four
already-ready observations. No waiting is added to fill groups. Group1 uses the
same new consumer and preparation order, distinguishing grouping from merely
moving original-record construction outside the guard. Every observation still
has all 50 candidates checked under the guard. No label or fitting is involved.

Three repetitions of small read-only/write cases under race, multi-record
validation tests and vet passed before the 12-arm frozen load experiment.

| Trial | Mode | Accepted / 192 | Drop / expired | Read p99 (ms) | Write p99 (ms) | Accepted-age p95 (ms) |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| 0 | Off | N/A | N/A | 30.886 | 21.415 | N/A |
| 0 | Batch | 173 | 19 / 0 | 32.148 | 21.730 | 380.953 |
| 0 | Group1 | 183 | 8 / 1 | 50.086 | 33.717 | 326.392 |
| 0 | Group4 | 192 | 0 / 0 | 23.504 | 24.624 | 104.947 |
| 1 | Off | N/A | N/A | 36.838 | 18.877 | N/A |
| 1 | Batch | 170 | 22 / 0 | 30.098 | 17.786 | 390.683 |
| 1 | Group1 | 173 | 19 / 0 | 30.187 | 19.767 | 384.500 |
| 1 | Group4 | 188 | 0 / 4 | 31.440 | 28.133 | 108.975 |
| 2 | Off | N/A | N/A | 32.121 | 18.980 | N/A |
| 2 | Batch | 175 | 17 / 0 | 31.173 | 19.819 | 359.733 |
| 2 | Group1 | 178 | 14 / 0 | 33.036 | 21.967 | 374.575 |
| 2 | Group4 | 192 | 0 / 0 | 28.845 | 20.024 | 90.156 |

Nearest-rank quantiles. All group4 age p95 values pass 250ms and all paired
read-p99/off ratios pass 1.10. Group1 does not pass age, so preparation order
alone does not explain the improvement. Group4 admitted 572/576 (99.31%),
validating 28,600 candidate records. There were zero queue drops. A single guard
entry deadline in trial 1 rejected an entire four-observation group; rejection
is deliberately not counted as learning or accepted work.

Guard acquisitions were 50/49/49, versus 184/173/178 for group1. Groups never
exceeded four, and partial groups were processed. This supports acquisition
amortization as the mechanism. However, group4 write p99 is higher than off in
all trials (about 1.15x/1.49x/1.06x). Grouping can transfer cost and amplify a
deadline failure to multiple observations. No write noninferiority guarantee,
durable-learning throughput, predictive gain or real-agent validation is shown.

Next run the unchanged 12-arm protocol again before considering further work.
Do not tune caps or discard the deadline failure. Full feedback/history authority,
durable commits and fitting remain outside this experiment.

## Artifact

`mmm-grouped-load-v27.jsonl`, SHA-256
`5b559d766fcc05293b8a2b0c34c76cee0f7ba2bf46a9ec05dcb5d5f524bb8051`.
Independent verification checked all 12 distinct cells, embedded source hashes,
192 recalls/96 writes per cell, group-weighted phase/outcome conservation,
50 validations per accepted observation and no execution errors. No production
configuration, deployment or push occurred.
