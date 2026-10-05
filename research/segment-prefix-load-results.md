# Cold prefix service: conditional gain, invalidation regression

**FAIL all six completion cells.** Cold preparation is inside the deadline;
no candidate state is prepared from reference forecasts or outside the job.

| Workload | Trial | Table completed | Prefix completed | Successful preparations | Reuse attempts |
|---|---:|---:|---:|---:|---:|
|Shared prefix|0|2|4|1|23|
|Shared prefix|1|2|5|1|25|
|Shared prefix|2|2|7|1|24|
|Earlier input changed|0|3|1|1|0|
|Earlier input changed|1|2|1|1|0|
|Earlier input changed|2|2|1|2|0|

All denominators64 including drops and stale work. Every returned forecast
matches its own reference; no request errors and16 overlapping writes per arm.
Two shared-prefix successful returns were rejected after processing, so verified
returns are not interchangeable with completed scheduler outcomes. Failed
preparation attempts are not included in successful-preparation counts.

Shared-prefix p99 passes paired1.10x limits in all trials. Invalidating trial1
and2 fail:30.008 vs25.080ms and30.982 vs26.800ms. Small paired results do not
establish population tail guarantees. There is no shadow-off arm here.

Zero reuse in the invalidation control is essential correctness, not a failure
to be bypassed. The extra cold preparation makes this path worse than rebuilding
directly when earlier evidence keeps changing. Retained prefix memory exceeds
5.5MB and failed preparations also allocate; actual peak RSS/GC attribution
was not measured, so memory-pressure causality remains unproven.

Artifacts: `segment-prefix-load-results.json`, verification via
`node research/verify-segment-prefix-load.mjs` (passed:false). Full per-request
timings, processor deadline traces and source hashes are retained.

## Next decision

Do not adopt this as the default full fitter. A future explicit selector would
need to justify expected reuse before paying preparation cost, without using
future request identities or oracle hit rates. That still would not solve the
original all-distinct capacity target. A separate incremental algorithm could
update arbitrary overlapping sufficient statistics, but must preserve the model
and as-of rules or declare approximation; this prototype does not do that.

The arithmetic and prepared-prefix leads are now bounded by real service
failures, not just component timing. Broader direction6 requires useful outcomes
and realistic arrival/load contracts alongside compute improvements. Production,
whitepaper claims and the seven open statuses remain unchanged.
