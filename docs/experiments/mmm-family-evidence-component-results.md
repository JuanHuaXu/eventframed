# Same-window family evidence: component ready, efficacy untested

This is a research adapter for an existing model, not a new family-learning
algorithm. family.go/family_prior.go already implement the same-window posterior.
The adapter reuses familyLogEvidence, records marginal evidence/posterior weights,
and omits unused partial-view tables. Its predictions match fitPriorConditional
for all512 full-input states at tested priors0,.5,.95,1.

The earlier V91/V92 static-data failures remain valid. V92 used generic prior.9;
the current stream protocol froze.95 before that historical comparison. No
quality result exists yet, and the prior must not be tuned to the coming run.

## Verification

Final reuse race run:45.330s package time. Stream contracts43.59s, reference
contracts.24s, existing family evidence test.04s. Four fixtures cover stationary
and reverse-transition cases under immediate and delayed/missing schedules,
both publication cadences and both32/64 windows. Original32-clock fit origins
match exactly. Four poisoned prefixes preserve fitted evidence and predictions;
input serialization is unchanged.

Component checks include independent Beta-integral evidence,20 predictive ratios,
all512 predictions against the existing no-boundary segment component, equality
with the existing parameterized-prior implementation, sample-order invariance,
prior endpoints and invalid inputs. The adapter computes both families from the
same sample slice; it never compares evidence from unequal window counts.

## Cost

Apple M4,20 iterations x3, no concurrent experiment. Final reused aggregation
costs15.47-16.43us,4864 bytes and one allocation. Full64-sample two-family fitting
costs7.29-7.56ms,353024 bytes and four allocations. These are small fixture
benchmarks, not tail-latency or whole-daemon throughput evidence.

The earlier local aggregation implementation measured30.44-31.03us. That output
is retained, but the final implementation uses the existing marginal-evidence
primitive. Neither benchmark includes retrieval, persistence or orchestration.

## Next step

Build the full four-variant collector and independent scorer, then run all2688
consumed trajectories against the matched-cadence Markov controls. No efficacy
collection has run yet. Do not promote this component from mathematical
agreement or speed alone. All seven full research goals remain open.

Artifacts: [protocol](mmm-family-evidence-protocol.md),
[final race](mmm-family-evidence-reuse-contracts.txt),
[final benchmark](mmm-family-evidence-reuse-benchmark.txt),
[earlier race](mmm-family-evidence-contracts.txt),
[earlier benchmark](mmm-family-evidence-benchmark.txt).
