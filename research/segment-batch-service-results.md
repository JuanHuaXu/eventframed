# Batch normalization: service completion unchanged

**FAIL all six completion cells.** Both pairwise and batch fitters complete
exactly1/64 in every trial, despite the separate27% fresh-fit component gain.
This result must not be described as a service rescue. No cache is involved.

| Work | Trial | Off p99 ms | Direct p99 ms | Batch p99 ms | Batch completed |
|---|---:|---:|---:|---:|---:|
|Grouped|0|32.017|30.834|31.621|1/64|
|Grouped|1|41.428|31.614|32.040|1/64|
|Grouped|2|31.679|32.005|31.503|1/64|
|Distinct|0|29.900|30.881|35.715|1/64|
|Distinct|1|30.784|31.558|27.660|1/64|
|Distinct|2|32.731|27.492|31.977|1/64|

All18 arms are error-free with16 overlapping writes each. Batch offered jobs
include19-21 drops and42-44 stale outcomes per arm, all retained in denominators.
All returned forecasts match their own reference within1e-12; zero cache hits.
Distinct trial0 violates the1.10x latency limit against both controls, and
distinct trial2 violates it against direct fitting. Small samples do not
establish population non-harm even for the other cells.

Evidence: `segment-batch-service-results.json` stores1152 read and288 write
durations plus status, verified-return counts and source hashes. Run
`node research/verify-segment-batch-service.mjs` for accounting/source checks
and the failed research verdict. The Go test passing means collection finished.

## Next causal diagnostic

Per-request latency and aggregate cancellation counters cannot locate the
remaining failure. Record remaining deadline at processor entry, actual fit
duration, cancellation and post-fit freshness separately. The offered load may
leave most jobs too little budget to finish even after arithmetic improvement;
this remains a hypothesis, not a demonstrated cause. A store-compatibility
check can also consume time around computation. Do not tune model accuracy,
raise deadlines or claim an incremental algorithm is necessary solely from
these aggregate counters. Production and all prior artifacts are unchanged.
