# Generic/Boolean family v91: frozen component protocol

Frozen before the new experiment. v90 remains FAIL; this is a different model
and a fresh experiment, not a reinterpretation of its threshold or results.

Use the coherent family mixture in research/boolean-family-rescue-proposal.md.
Family priors are equal. Each family's subset prior is inclusion1/3, with the
unchanged Beta(1/2,1/2) parameter priors and sequence likelihoods. Both use the
same uniform input law. Family evidence is summed over alternatives, never
multiplied as independent evidence sources. Compile one immutable predictive
law. No production path, observer, journal or residual certificate changes.

Keep v90's ten cases,64 fits per phase/case, nested16/32/64 samples, disjoint
phase-specific rule pools, cost6-feasible generator masks, and exact scoring
over512 raw inputs. Views remain full and fixed mask63. Control remains the
unchanged subset model. Candidate is the family mixture, not the Boolean-only
model. Store family weights in addition to training/prediction hashes and
scores. These are posterior probabilities under the declared finite family,
not universal confidence that a real-world rule is correct.

The seed formula changes to2030119100 + phase*1000000 + case*10000 + index*10
+ role, with role0 mask,1 inputs,2 label noise. Verify no role seed overlaps any
v90 fit seed. Rule pools are the same specification as v90, so they are not new
unseen domains. Fresh fitting draws provide new evidence for the new model.

## Frozen advance rule

For parity3, parity4 and complemented parity4 at n32 and n64, in BOTH phases:
for each paired fit let g=L_control-L_candidate and e=L_control-L_oracle, using
full-input expected Brier. Require the mean of g-0.2e to have a positive lower
mean-minus3.5SE bound. This tests removal of more than20% of control excess
risk, including uncertainty, instead of demanding more than available headroom.
The exact simulator oracle is evaluator-only and never provided to fitting.

For ALL ten cases, both phases, all three sample counts, and both views require
the paired upper mean-plus3.5SE bound on candidate-minus-control Brier <=0.01.
Report every observed regression, including those inside tolerance. Thus small
sample counts and non-parity behavior can veto advancement. No majority/mux
exception. All required conditions must hold; passing some does not pass all.

The bounds are approximate across-fit normal bounds, not anytime confidence
sequences. Nested sample sizes are not independent extra fits. Passing advances
only to a separately frozen delayed-stream integration; it cannot satisfy v88
or a whole research direction. No threshold changes after seeing these data.

## Verification

Independent Beta-integral family evidence, normalized mixture, full and partial
mixture identity, sample-order/complement/permutation symmetry, immutable reads,
input bounds and null controls. Same sequence is used once per alternative
model family; normalization is not a product of family evidences.
Preserve all frozen predecessor sources and verify exact v90 replay separately.
Store all3840 records and source/protocol hashes, replay exactly, run race/vet,
and measure same-fixture fit/lookup cost with allocations. Delayed eligibility,
input-law misspecification and real-world serving remain untested here.
