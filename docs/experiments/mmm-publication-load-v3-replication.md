# Unchanged publication workload replication

All embedded v3 sources matched the working files before execution. The same
frozen test ran again into a new exclusive artifact; no parameters, thresholds,
labels or implementation were retuned. This is a repeated performance workload,
not independent task/model generalization evidence.

[Raw replication](mmm-publication-load-v3-replication.jsonl), queue16:

| Requests | Trial | Completed | Age p95 ms | Serving p99 on/off |
| --- | ---: | ---: | ---: | ---: |
| 64 | 0 | 56/64 | 101.526 | .399 |
| 64 | 1 | 55/64 | 101.164 | .733 |
| 64 | 2 | 61/64 | 88.839 | .872 |
| 192 | 0 | 154/192 | 106.235 | .786 |
| 192 | 1 | 158/192 | 106.173 | .872 |
| 192 | 2 | 155/192 | 105.254 | .879 |

Queue16 PASSED all six finite cells again. Longer completion margins remain
narrow (80.21%-82.29% versus80% required), even though the age margin is larger.
All admitted50-label frontiers completed with zero reported errors.

Queue64 again completed every request but FAILED two cells:64-request trial1
serving p99 was1.251 times off (>1.10), and192-request trial1 age p95 was265.070ms
(>250ms). The overall test therefore correctly failed; repeated complete delivery
does not establish timely learning or stable serving tails for this buffer size.

Source hashes,18 arm records, request/write counts, label/frontier conservation,
age cardinality and individual gates were audited. Both original and repeated
outcomes remain recorded. Timing ratios do not prove a population speedup;
sequential arm scheduling and shared-host noise remain limitations.

## Research decision

Queue16 remains the candidate for a fully covered adapter, not a production
setting change. The current test adapter intercepts only ingestion. Required
general integration includes policy binding, composition, deletion/retention,
selection/AP/omission certificates, Bayesian outcomes, predictive snap/rollback,
agency proposal/claim/resolution, maintenance and recovery. Each must either
participate in coherent publication or explicitly disallow fast proofs while
it can affect state. Embedded unwrapped mutation methods cannot be assumed safe.

No whitepaper claim of general readiness, production change or push was made.
Other research directions, real-agent outcomes and persisted feedback recovery
remain open.
