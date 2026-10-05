# Boolean specialist v90: frozen component protocol

Status: frozen before running the component experiment. This follows the v89
diagnosis, not a change to v88 thresholds or a rescue of its failed runs.

## Boundary and model

Implement the joint model in `research/boolean-specialist-proposal.md` with
independent inclusion prior1/3 over all512 subsets and Beta(1/2,1/2) agreement
noise. Fit1..256 supplied labeled samples. Compile predictions into the existing
immutable ConditionalForest with uniform independent input law. No generator
mask, scenario name, current query or future outcome is a fitting argument.
No production integration, observer change, extra audit, or selector change.

Enumerating512 rules is exponential in dimension; nine bits is the explicit
research cap. A noisy parity specialist is not a universal Boolean learner.
The uniform input assumption is deliberately misspecified in one control.

## Component design

Two phases,64 independent fit replicates per case, and nested sample sizes
16,32,64. Cases: parity1, parity2, parity3, parity4, complemented parity4,
majority3, multiplexer3, constant-zero, null, dependent parity4. All non-null
rules receive independent5% label flips; null labels are independent fair coins.
Dependent inputs set coordinate8 equal to coordinate1 before labeling.

For each arity, enumerate masks whose nested-view acquisition cost is at most6,
including mandatory coordinate0. Split the sorted mask list by index parity:
even indices for design, odd for confirmation. Select uniformly within the
phase pool. The specialist's prior covers all masks in both phases; there is no
privileged target mask. These are disjoint generator-mask pools, not an unseen
domain claim. Constant/null have no rule mask. Multiplexer uses the lowest
selected coordinate as selector, next as true branch, last as false branch.

Use standard Go math/rand with seed
2026119000 + phase*1000000 + case*10000 + replicate*10 + role,
where phase0/1, case0..9, replicate0..63; role0 chooses mask, role1 draws
training inputs, role2 draws training label noise. Training prefixes are paired
across sample sizes; arms receive identical samples.

Compare new Boolean specialist and unchanged NewSubsetConditional. Score each
fitted law over all512 raw inputs, weighting each by1/512 before the dependent
input mapping. For each input, use the simulator's conditional probability of
Y=1 only in the evaluator. Expected Brier is
(p-p_true)^2+p_true*(1-p_true). Expected accuracy uses threshold p>=1/2.
This exactly integrates fresh outcomes under the finite simulator; no sampled
test-label uncertainty is implied. Variation comes from independent fitted
training sets. Do not describe these rates as observed chatbot accuracy.

Report full-input forecasts and a paired fixed observed mask63 (coordinates0..5,
cost6). The fixed mask is independent of the target and shared by both arms;
it is not an adaptive MMM observer. Full-input quality is a model diagnostic,
not a charged foreground retrieval policy. Score oracle conditional risks for
each observed mask as lower bounds under this simulator.

## Advance rule, not adoption

At n32 and n64, require mean full-input Brier gain at least0.01 with positive
mean-minus3.5-standard-error bound for parity3, parity4 and complemented parity4
in BOTH phases. Report all other cases, n16, partial forecasts, and all observed
regressions. These are approximate across-fit normal bounds, not confidence
sequences or an anytime certificate. Passing advances the specialist to a
separately frozen delayed-stream integration experiment. It does NOT pass v88,
establish all-case non-inferiority, or satisfy any whole research direction.

Non-parity degradation falsifies replacement of the generic learner, even if
the narrow advance rule passes. Future integration must retain the generic
learners and meet unchanged whole-stream harm and recovery gates.

## Verification and cost

Verify Beta sequence evidence without a binomial coefficient, normalized
weights, agreement/complement symmetry, input-permutation equivariance,
order invariance, analytic partial marginalization and immutable snapshots.
Reject empty/oversized/out-of-range samples. Test constant and null controls.
Timestamp/selection eligibility remains the caller's responsibility: this
component cannot certify that supplied labels were available at prediction time.

Store all3840 per-fit records and hashes of predictions, training data, sources
and protocol. Replay exactly. Run focused race tests, vet, and a narrow audit.
Measure fit and compiled forecast cost with allocations. Report machine/runtime
and component-only scope; do not infer daemon tail latency or asynchronous
integration safety from microbenchmarks. No edits to frozen predecessor sources.
