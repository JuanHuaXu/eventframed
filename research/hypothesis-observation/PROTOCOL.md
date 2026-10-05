# Explicit hypothesis observation v1

Frozen 2026-09-12 before results. Roadmap7 first mechanism experiment. This is
a bounded diagnostic hypothesis family, not a full EventFrame serving path.

H=(target bits0,1; nuisance bits2,3) has16 equally likely states. The scored
class is the two target bits (four classes). Eight unit-cost binary tests:
target bit0, target bit1, nuisance bit2, nuisance bit3, target parity, nuisance
parity, fair coin, target0 XOR nuisance2. Bit tests use declared noise n for
target, .01 nuisance, min(.4,2n) target parity, .05 nuisance parity, .5 coin,
and n mixed. Test outcomes are conditionally independent given H in matched
cases, including repeated queries. Sixteen acquisitions per episode.

Four policies, same likelihood model and prior: uniform random test; maximum
predictive label entropy; maximum expected information gain about full H;
maximum expected reduction of target-class Gini impurity. The last explicitly
reduces posterior mass on competing target explanations rather than spending
queries distinguishing nuisance states. It is an expected class-separation
heuristic, NOT a faithful EC2 implementation or its approximation guarantee.
Source: Golovin, Krause & Ray,
[Near-Optimal Bayesian Active Learning with Noisy Observations](https://arxiv.org/html/1010.3091),
Sections3/4. Their distinction between hypotheses and equivalence classes
motivates the objective; our test bank, noisy resampling and Gini rule are ours.

Cases: noise05; noise20; all_relevant (all16 H are distinct target classes);
null (every likelihood .5); correlated (noise05 marginals but each test's first
response is reused, violating the learner's independent-repeats likelihood).
All_relevant removes nuisance distinction. Null is an unidentifiability control.
Correlated is an explicit model-misspecification stress case, not a claimed
valid posterior or certified operating domain.

256 episodes/case/split, two disjoint splits. Seed=2026100201*1000000 +
split*100000 + case*1000 + episode; role substreams derived by seed*10+role,
role0 truth,1 outcome tape,2 random policy. Every policy gets the same latent
state and a common pre-generated16x8 uniform outcome tape. Policy selection
receives only posterior and likelihood table, NEVER truth or future tape.
Correlated uses tape[0][test] for every repeated result. No fitting/tuning
between design and confirmation. Labels are generated, not public real facts.

Journal pre-query class probabilities and selected test before outcome, then
update posterior from the selected likelihood. Brier is sum over class squared
errors; mean over16 pre-query forecasts measures learning speed. Also report
final posterior accuracy/Brier and confident-wrong final decisions (max>=.9).

Conditional mechanism success: on BOTH noise05/noise20, target-Gini must improve
mean learning-curve Brier by>=.02 with paired z=3.3 lower>0 against random AND
label-entropy. All_relevant/null mean harm vs random must be<=.01. Full-H
information gain is an additional strong comparator, not a predeclared required
win. Report all cases and controls; no general-use pass if correlated high
confidence errors remain. A matched-model pass does not complete roadmap7.

Verify exact posterior normalization, fair-coin no-update, selection independence
from hidden truth, pre-outcome journal reconstruction and full deterministic
replay. Keep failed cases, including model misspecification, visible.
