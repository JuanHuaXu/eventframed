# Exact reuse: repeated-work pilot passes

**PASS the frozen diagnostic in all three pairs.** This is a conditional rescue
for identical fitting inputs, not completion of research direction6.

| Trial | Direct p99 ms | Memo p99 ms | Direct completed | Memo completed | Memo drops |
|---|---:|---:|---:|---:|---:|
|0|29.340|27.926|1/64|58/64|6|
|1|30.379|27.499|1/64|60/64|4|
|2|39.592|30.159|1/64|61/64|3|

All16 writes overlapped reading in all six arms. Zero read/write errors. Every
offered job is accounted for; memo has zero stale/interrupted jobs. Direct
controls have40/40/41 stale jobs and23/23/22 drops. Cached completion is not a
new fit or additional observation: repeated calls reuse an identical computed
predictive vector, retaining per-request deadline and snapshot checks.

The actual scheduler, Recall/Observe and temporary persistent LibraVDB are
used. Each arm starts with a cold one-entry cache; the external warmup does not
populate it. Exact-key control tests passed under race instrumentation for
tenant, epoch, evidence outcome/input/availability, cap, hazard, family mass and
clock/length changes. Canceled work cannot use a hit or publish a new result.
The cache is test-only and process-local; algorithm version is fixed by its
test binary, not a general persisted cache-version contract.

Raw artifact `segment-transform-memo-results.json` contains request timings,
terminal counts and source hashes. Verify via
`node research/verify-segment-transform-memo.mjs`. The scientific screen passes,
not just the experiment's successful execution. This is a small diagnostic
sample without population tail confidence bounds.

## Limits and next experiment

The processor deliberately fits the same labelled fixture for every request;
it does not learn from these recall results. This is highly favorable to exact
reuse, so it demonstrates a mechanism rather than an expected real-world gain.
The remaining3-6 dropped jobs stay in the denominator. Neither packet accuracy
nor real agent usefulness is measured. Production is unchanged.

Before adoption, vary work identity in the service workload (not only unit
tests): all-distinct histories, intermittent new evidence, and repeated work
with obsolete dependencies. Check cache misses, no cross-history predictions,
freshness, foreground latency and offered-job completion together. Compare to
shadow-off as well as direct fitting; today's two arms both run shadow work.
Do not pool this positive fixture with failed earlier workloads to claim an
unconditional rescue or relax their thresholds.
