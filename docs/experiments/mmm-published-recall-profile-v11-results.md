# Full Recall admission profile v11: capacity diagnosis, not a rescue

Date: 2026-10-02. The [frozen protocol](mmm-published-recall-profile-v11-protocol.md)
profiled the unchanged v9 and v10 tests once each in normal mode. Both
retained their expected failing service results. This is diagnostic evidence,
not a new Goal 6 pass. Git HEAD was
`1a7edb62b6be4031fd01ebeab8071b17303a7815`; Go was `go1.27.1`
on `darwin/arm64`. The test and protocol SHA-256 values are recorded in the
[v9](mmm-published-recall-load-v9-results.md) and
[v10](mmm-published-recall-admission-v10-results.md) reports; v11 protocol
SHA-256 is `3def155fd7b1b8858a803dbdc897060779667d4d4cb51f1d18c2176b4f7770aa`.

| Profiled trial | Writes | Successful Recalls | Stale rejections | Recall call p99 | Offer-to-done p99 |
| --- | ---: | ---: | ---: | ---: | ---: |
| v9, no admission | 128/128 | 117/128 | 87 | 101.547 ms | 529.249 ms |
| v10, admission | 128/128 | 128/128 | 0 | 63.170 ms | 526.315 ms |

The unchanged second internal trial had v9 118/128 Recalls, 87 stale
rejections and 580.333 ms offer p99; v10 had 128/128 Recalls, zero stale
and 531.856 ms offer p99. These profiled timings are not unprofiled
performance estimates. The v10 writer offer-to-ack p99 was 85.869/83.553 ms;
published-view maximum age was 32.102/32.204 ms.

The Go execution-trace synchronization profiles attribute aggregate
goroutine blocking, **not per-Recall latency**. In v10, cumulative blocking
under `GetResidualCandidates` was 3.138 s and under the adapter's
`PutBayesianJournal` was 4.080 s across both internal trials. The
corresponding v9 spans were 3.145 s and 4.906 s. Many other waits are
test channels, database connection maintenance and admission; summing them
as request time would be wrong. V10 CPU samples were dominated by Go
scheduling/wakeup routines under this instrumented run, with HNSW insertion
also present; the CPU profile alone does not locate the critical path.

The source reveals an independently testable hot-path cost: each successful
Recall nominates 150 candidates, `loadResidualCandidates` calls
`GetResidualCandidates` once per candidate with eight workers, and each
call does two point reads (exact and general). For 256 successful v10
Recalls, the upper count is 38,400 candidate API calls and 76,800
residual-record point reads. This fixture creates no residual records.
The count does not prove those reads caused the queue; the journal owner,
durable insert, HNSW writes and RW admission are competing causes.

Ranked next checks:

1. Remove empty-residual lookup work in a *test-only* matched ablation.
   Falsifier: occupied-call and offer-to-done tails do not materially
   improve. This cannot justify disabling residuals in general.
2. If capacity remains low, isolate journal owner wait and durable insert
   from read admission and HNSW write contention with per-phase spans.
3. Only then test a general residual-batch or version-certified empty-cache
   design, including nonempty and concurrent-residual-change controls.

The profiles and traces remain under `/tmp/eventframe-goal6-profile-v11/`,
outside the repository. They may be removed by ordinary temporary-file
cleanup; the source, protocol and result are the reproducible checkpoint.
No production behavior changed. All seven whole goals remain open.
