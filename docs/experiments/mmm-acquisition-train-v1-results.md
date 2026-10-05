# Full acquisition-to-training result

Status: FAIL added-value criteria.504/504 non-harm checks pass;0/48 required
improvements pass. All2688 consumed trajectories retained, both schedules and
all21 cases. Not fresh confirmation or a production learning upgrade.

Phase1 delayed terminal64 served-mixer Brier:

| Case | Natural | Random | Entropy | Disagreement |
| --- | ---: | ---: | ---: | ---: |
| Additive stationary | .221755 | .221961 | .220217 | .220710 |
| Parity4 | .049066 | .048874 | .048661 | .048812 |
| Null | .258525 | .256671 | .258779 | .258398 |
| Majority to parity | .060262 | .055775 | .056636 | .056180 |
| Parity to majority | .102272 | .098198 | .098686 | .096202 |

Each paid policy uses31 queries in these cells. Parity-to-majority disagreement
gain over natural is.006071, interval[-.001890,.014032]; over random it is.001997,
interval[-.009348,.013341]. Neither supports the required reliable improvement.
Intervals are exploratory paired mean +/-3.5SE, not simultaneous guarantees.

Individual expert means also change: in that cell generic64 natural.108775
becomes.100192 under disagreement; generic32.102642 becomes.099902. This confirms
the intervention reaches fitting, not just mixing. It does not establish robust
acquisition superiority. All per-expert Brier/accuracy/log-loss arrays are kept.

## Verification

The four-worker collector completed in340.31s test time,343.16s shell wall time.
Raw forecasts, fits, queries, source hashes and data hash are retained. Natural
control matches original four experts and Markov on every run. No record errors,
unequal query budgets or immediate-schedule prediction changes occurred.

Independent JS scorer checked13,762,560 forecast values, all fit-origin lists,
query reveal rules and metrics; maximum metric difference from Go4.44e-16.
Scorer replay is byte-identical (`cmp` exit0). This is a scoring replay, NOT a
second full Go collection. Prior component tests passed under race; the complete
quality collection itself was not race-instrumented. Offline runtime is not
serving latency or a production throughput benchmark.

Initial scoring attempt failed because a hashing data listener started one file
flowing before its line iterator attached. Fixed by attaching both iterators
before the first await. The immutable collected artifact was not altered or
regenerated; complete alignment/count/hash checks now pass. No global rule changed.

## Next lead

Before increasing budgets or changing learners, measure how many paid labels
actually enter a fit EARLIER than they would have naturally. Acquisition can
improve mixer feedback yet provide no new training information if the original
label arrives before the next32-frame publication. Queries scheduled just after
publication are especially exposed. Quantify that mechanism with the retained
query/fit records, then test timing alignment with equal budgets if warranted.
Do not assume that every paid query is an additional effective training example.
All seven full research directions remain open.

Artifacts: [protocol](mmm-acquisition-train-protocol.md),
[raw](mmm-acquisition-train-v1.jsonl), [run log](mmm-acquisition-train-v1-run.txt),
[summary](mmm-acquisition-train-v1-summary.json),
[scoring replay](mmm-acquisition-train-v1-summary-replay.json).
