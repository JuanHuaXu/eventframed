# Bound worker v12: held source proof and rate matrix

Date: 2026-10-01. [Frozen contract](mmm-bound-worker-guard-v12-contract.md).
Research-only; production unchanged. Apple M4, Darwin arm64, ordinary Go
build for latency. The full Recall path used a hash embedder, one visible
synthetic event, four probe workers, 192 paced offers per trial, 64 guarded
labels per trial, and 256 future-only writes in the writer arm. Three paired
trials were run at each rate, with arm order alternated. Each arm/rate pooled
576 Recall calls and 192 labels; each writer rate completed 768 writes.

The new research guard waits with a deadline and holds **one** writer permit
through as-of validity, source lineage, journal/feature/witness validation,
and durable admission or feedback. It replaces the bound worker's two-step
fail-fast source check. Unit controls confirm that the callback does not run
while another writer holds the permit or after deadline expiry; future-only
motion passes, while source deletion and backfill reject. Existing legacy
fail-fast APIs were not changed. Focused race tests, full affected-package
suites, and vet pass.

The table shows the first full matrix run. After the final post-acquisition
lineage-fault check was added, a second complete matrix run preserved the same
outcome: writer offer p99 was 877.39, 627.74, 267.92, and 32.38 ms at 1, 2,
4, and 8 ms intervals respectively. Writer label-age p99 was 24.91, 20.45,
21.89, and 23.95 ms. Thus the 4 ms failure and 8 ms pass both replicated on
the final code state; these are repeated local finite runs, not an independent
task or hardware population.

| Offer interval | Quiet Recall offer p99 | Writer Recall offer p99 | Writer call p99 | Writer queue p99 | Writer label-age p99 | Frozen writer gate |
| --- | ---: | ---: | ---: | ---: | ---: | --- |
| 1 ms | 377.03 ms | 851.67 ms | 41.00 ms | 829.79 ms | 22.70 ms | diagnostic overload |
| 2 ms | 179.47 ms | 650.61 ms | 38.08 ms | 623.36 ms | 23.25 ms | diagnostic overload |
| 4 ms | 11.93 ms | 247.29 ms | 37.11 ms | 225.71 ms | 22.41 ms | **FAIL** Recall <100 ms |
| 8 ms | 9.96 ms | 30.03 ms | 30.03 ms | 0.065 ms | 20.31 ms | **PASS** both bounds |

All 24 trials completed 64 labels without failure, retained source-bound
durable records, and reopened their motion workers after the workload. Writer
trials moved the backend runtime version from 2 to 258 and recorded active
learner/write overlap. No future event entered the as-of frontier. The 8 ms
cell supports a finite ~125 Recall offers/s point on this fixture; the 4 ms
cell fails at ~250 offers/s. The 1 ms failure from [v11](mmm-bound-worker-load-v11-results.md)
remains valid; v12 fixes its guard-busy loss, not its overloaded request queue.

Reproduce:

```sh
EVENTFRAME_RUN_MOTION_RATE_V12=1 EVENTFRAME_RESEARCH_PERF_GATE=1 go test ./internal/service -run '^TestResearchMotionBoundWorkerRateMatrixV12$' -count=1 -v
go test -race ./internal/researchpublicationstore ./internal/service -run 'TestResearchSourceAsOfGuard|TestResearchMotionBoundWorkerFittedAbruptExit|TestResearchMotionBoundWorkerAcknowledgedAbruptExit|TestResearchMotionBoundWorkerFutureOnlyRestartAndBackfill' -count=1
go test ./internal/researchpublicationstore ./internal/service -count=1
go vet ./internal/researchpublicationstore ./internal/service
```

The first command intentionally exits nonzero because the frozen 4 ms gate
fails after reporting every matrix cell. This matrix was designed after
observing v11's saturation and is not untouched confirmation. It is not an
OpenClaw/LLM workload, a large-corpus result, a population p99 guarantee,
hardware power-loss proof, or evidence of better agent answers. Goal 6 OPEN.
