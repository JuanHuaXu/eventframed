# Empirical input integration: rejected

320 fresh streams with independently fitted4096-label bases, ten cases and
two16-trajectory cohorts. The frozen candidate changes only input weights for
the retained subset model, using its same64 past audit samples plus one total
uniform pseudocount. Anti-Pigeon decisions, labels, fitting cadence, mixture
updates and split behavior remain unchanged.

Two candidates: fixed to the control's actual observations, and coupled to its
own observer with the same six-coordinate cap. No oracle input distribution.

## Quality

Post-period scalar Brier in the second cohort (lower is better):

| Case | Uniform control | Empirical fixed observation | Empirical coupled |
|---|---:|---:|---:|
| Bit, uniform | .202465 | .202996 | .203773 |
| XOR2, uniform | .206934 | .207630 | .208031 |
| Majority3, uniform | .222112 | .221449 | .221351 |
| Multiplexer, uniform | .222756 | .222646 | .223009 |
| Parity4, uniform | .238169 | .240566 | .253799 |
| Bit, dependent | .199606 | .198802 | .199467 |
| XOR2, dependent | .217384 | .214182 | .215078 |

No coupled changed-case cell passes the frozen .005 mean gain plus positive
mean-minus3.5SE gate in either cohort. Parity4 fails the .01 non-harm screen in
both cohorts: mean degradation .010879 (design) and .015631 (second).
Second-cohort paired gain interval is [-.031010,-.000252]. All other coupled
non-harm screens pass; stable/null effects remain tiny. These are exploratory
trajectory intervals, not anytime or universal coverage guarantees.

Dependent gains do not reproduce the magnitude of the fixed-mask component
probe. Second-cohort coupled gains are .000139 for bit and .002306 for XOR2,
both intervals crossing zero. Fixed-observation gains are .000804 and .003203.

## Interpretation

Reject unconditional empirical-input replacement. The same frozen estimator
that resolves a deliberately selected partial-input dependency can hurt an
adaptive observer. The parity4 fixed-arm degradation is .001603/.002397,
whereas coupled degradation is .010879/.015631. This controlled contrast points
to the observation/forecast feedback loop as an additional failure mechanism;
it does not identify which acquisition decision first causes the harm.

Next useful evidence is a pre-label paired path trace for parity4 and the
dependent cases, with equal-budget information sufficiency and stopping
decisions checked. Inspect earlier acquisition diagnostics first. Do not tune
pseudocounts on these consumed cohorts or substitute fixed-arm benefit for
primary success. The older v11 forest failure remains unchanged.

## Cost and Verification

Fixed-arm acquired coordinates exactly equal control at full/post levels.
Coupled mean full-stream costs often decrease slightly. Second parity4:
4.64819->4.63098 coordinates/frame; dependent bit:4.68835->4.61853;
dependent XOR2:4.63818->4.57947. Less acquisition is not success when forecasts
degrade. Audit, training and split schedules are identical. This does not
establish equal total compute or serving-latency improvement.

- Original-control metric/tape parity on uniform, dependent-shift and stable
  dependent cases passed under race detector,3.658s; package vet passed.
- Experiment completed all320 records in60.46s, including repeated fits and
  artifact generation. No serving-latency interpretation.
- 816 captured source hashes verified against both embedded and current files.
- Unique fit seeds, cell counts, shared split/fit counts, bounded acquisition,
  fixed-arm exact costs and finite Brier metrics verified.
- Summary replay byte-identical. No performance campaign for the rejected
  candidate; no runtime, production, remote or whitepaper changes.

Raw:`mmm-subset-input-integration-v1.jsonl`.
SHA256:`24a5a70f080f92a40bd51308a23b051dfec3cec0ef824e0bd2dfede4929749ee`.
Frozen contract:`mmm-subset-input-integration-v1-contract.md`.
Evaluator:`research/subset-input-integration-summary.mjs`.
All seven research goals remain open.
