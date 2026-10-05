# Empty-residual Recall ablation v12: queueing remains

Date: 2026-10-02. The [frozen protocol](mmm-published-recall-residual-ablation-v12-protocol.md)
disabled residual lookup only in a test-only service with a verified
zero-residual fixture. The fixture had no `residual` records before
offers, and `ResidualVersion` did not change during either trial.
The v10 control was rerun immediately afterward with the original
configuration and unchanged source. Both arms retained the same
128 writes and 128 Recalls at nominal 4 ms offers, four Recall workers,
read-to-journal admission, exact published-LSN top-150 oracle and
durable journals.

| Metric | v12 no residual lookup, trials 1/2 | Fresh v10 control, trials 1/2 | Gate |
| --- | ---: | ---: | ---: |
| Durable writes | 128/128, 128/128 | 128/128, 128/128 | 128/128 |
| Successful Recalls | 128/128, 128/128 | 128/128, 128/128 | 128/128 |
| Stale rejections | 0, 0 | 0, 0 | 0 |
| Recall call p50 | 26.089, 26.079 ms | 27.112, 27.048 ms | Report |
| Recall call p99 | 62.708, 64.018 ms | 58.151, 64.081 ms | <100 ms |
| Recall offer-to-done p99 | **502.075, 503.120 ms** | **521.006, 528.993 ms** | **<100 ms** |
| Write offer-to-ack p99 | 85.837, 88.057 ms | 87.112, 86.855 ms | <250 ms |
| Published-view max age | 31.469, 31.284 ms | 35.415, 33.422 ms | <250 ms |

Both v12 trials passed all count, snapshot, future-leak, journal and
oracle checks; their only violation was the unchanged response-latency
gate. The ablation reduced median occupied-call time by about 1 ms
and offer p99 by 19/26 ms in these two matched executions. Because
these are only two runs under scheduler and write-load variation,
those differences are descriptive, not a confidence interval or a
general residual-lookup speedup. The central conclusion is robust to
that uncertainty: removing empty-residual reads did not clear the
roughly half-second queue or achieve <100 ms. Do not disable residuals
in a system with actual corrections.

The predeclared `-race` correctness-only mode passed two additional
v12 trials with 128/128 writes and Recalls and zero stale or semantic
violations. Instrumented timings were not compared with the normal
gate. Ordinary package tests and `go vet` passed.

Next falsifier: collect per-Recall elapsed spans for Search, graph and
certificate reads, residual phase, journal-owner wait, durable journal
insert/publication, admission wait and remaining service work on the
unchanged workload. If journal and write admission are the dominant
serial sections, test a durable ordering architecture that shortens
those sections without acknowledging before commit or loosening
as-of freshness. Do not infer this cause from the aggregate v11
blocking profile alone. Goal 6 and all seven whole goals remain open;
production is untouched.

Reproduce:

```sh
EVENTFRAME_RUN_PUBLISHED_RECALL_LOAD_V12=1 go test ./internal/store/libravdbstore -run '^TestResearchPublishedRecallLoadV12$' -count=1 -v -timeout 5m
EVENTFRAME_RUN_PUBLISHED_RECALL_LOAD_V10=1 go test ./internal/store/libravdbstore -run '^TestResearchPublishedRecallLoadV10$' -count=1 -v -timeout 5m
EVENTFRAME_RUN_PUBLISHED_RECALL_LOAD_V12=1 EVENTFRAME_RACE_CORRECTNESS_ONLY=1 go test -race ./internal/store/libravdbstore -run '^TestResearchPublishedRecallLoadV12$' -count=1 -v -timeout 5m
go test ./internal/store/libravdbstore -count=1 -timeout 5m
go vet ./internal/store/libravdbstore
```

SHA-256: v12 protocol
`401f22fbb6f9f2e351a48492568906a6f4c7538e81dc5993f37721b3f782c5c2`;
v12 test `702af5c5641541371610fd131c35177226838941bcfdc8a99a7b724bc4c5fc47`.
