# Temporal reuse does not rescue this workload

**FAIL all three completion cells; one latency cell also fails.** Both arms
run the actual scheduler and fitter, differing only in the existing temporal
compatibility policy. Strict matching completed1/64 in every trial; temporal
reuse completed1/64,0/64,1/64. The target remains52/64, not admission alone.

| Trial | Strict p99 ms | Temporal p99 ms | Temporal started/interrupted | Temporal dropped |
|---|---:|---:|---:|---:|
|0|30.990|34.562|22/21|23|
|1|31.004|25.797|21/20|21|
|2|36.990|31.581|20/19|24|

Strict started/interrupted counts were6/5,11/10,10/9. All arms have zero
request errors and16 overlapping writes; all64 offered jobs are accounted for.
Temporal accepted41/43/40; the rest were dropped. No failed/cancelled terminals.
One non-interrupted temporal job still became stale in trial1; cancellation
counts alone cannot describe every post-fit freshness failure.

The existing temporal guard admits more work across future-dated writes, but
most started fits exhaust their context before returning and queued work fills
capacity. This establishes that avoiding snapshot mismatch alone is not a
sufficient rescue. It does not prove writes are irrelevant or identify an
optimal worker count. No deadlines, guards or production defaults changed.

Raw per-request timings, terminal states, started/interrupted counts and source
hashes: `segment-transform-temporal-results.json`. Verify with
`node research/verify-segment-transform-temporal.mjs`; verdict is false.
Three small paired trials are diagnostic, not a population latency guarantee.

## Next lead

Investigate useful work admission: a fit should not start with less remaining
time than its declared cost allowance, and repeated equivalent requests may
join an in-flight computation only with an exact dependency key and their own
freshness check. Dropping expired work alone cannot satisfy the completion
denominator; report all offered jobs and distinguish shared valid results from
fresh fits. Do not share across differing labels, horizons, epochs or as-of
dependencies. The fixed fixture makes repeated computation especially redundant,
so any sharing pilot must include distinct-work negative controls before claiming
general usefulness. Actual real-task improvement remains untested.
