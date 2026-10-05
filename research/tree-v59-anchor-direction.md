# V59 prospective lead: baseline-anchored conditional tree

RECOMMENDATION, not a confirmed production bug or empirical rescue. V58 is
fully rejected on scientific gates; retain its source and all negative data.

## Gate And Competing Causes

Observed V58 risk rises from V57's .2178 to .2401; stationary risk rises about
.028. Candidate causes: (1) absolute-rate pooling destroys known baseline
differences, (2) only600 issued positions retain too little member evidence,
(3) measurement/source or latent-tree mismatch. No claim that (1) alone proves
the entire regression. A direct zero-evidence counterexample separates (1)
from delayed responses and noise; a full matched study must test its repair.
Falsifier: preserving the baseline conditional law fails to improve broad
quality/recovery, or stationary/cost protection still fails. No gate relaxation.

[Kull, Silva Filho and Flach (2017)](https://proceedings.mlr.press/v54/kull17a/kull17a.pdf),
section3.2 and propositions1-2, give probability calibration families containing
the identity and express them using log-score features. The methodological point
is to retain a correct incumbent possibility. Our finite paired-noise tree and
moment-anchored prior below are OUR construction, not their MLE algorithm,
empirical result, guarantee, or exact beta-calibration family.

## Declared Construction To Test

Retain the same public hierarchy, paired single-Y measurement model and suffix
initially. Replace a shared absolute rate with a shared conditional hypothesis
over each member's PUBLIC base b_i. Hypothesis0 is exactly p_i(0)=b_i, with
prior weight .8. The other20 hypotheses have weights .01 and offsets
d_j in {-.6,-1.2,...,-6,+.6,+1.2,...,+6}.

For each public b, choose a unique finite a_b solving

```
(1/20) sum_j sigmoid(a_b+d_j) = b.
p_i(j) = sigmoid(a_(b_i)+d_j), j=1,...,20.
```

Monotonicity and endpoint limits give existence and uniqueness for b in(0,1).
This calibration of a_b uses ONLY b and the frozen prior, never outcomes.
It is not a fitted reliability estimate. Cache all member/hypothesis rates.
The same common hypothesis index can be pooled; conditional rates differ by
public context. For every member, the prior predictive equals b exactly:
.8*b+.01*sum_j p_i(j)=b. Every zero-evidence tree pruning therefore preserves
the baseline, while hypothesis0 represents a correct stationary baseline
for every member. This is NOT a stationary non-harm theorem after updating.

Replacing rate means is insufficient: update paired likelihoods using p_i(j),
and evaluate each terminal conditional mean separately for its target member.
Do not retain V58's absolute-rate leaf means, six-category context-free
likelihood cache, or borrowed node-prior interpretation.

For bounded incremental computation, maintain finite log-factor sums and
integer zero-support counts for each(node,noise,hypothesis). Paired disagreement
under eta0 has zero likelihood; no subtraction of negative infinity. Exact
eviction removes the original member-specific factor. Failed normalization
must not commit trial or node state. Independent full-ledger reconstruction,
small-tree enumeration and original-factor conditionals are required before
cohort dispatch. Approximate O(depth*noise*hypotheses) update; forecast adds
O(depth*noise*hypotheses) instead of V58's cheap scalar leaf mean. Measure all
candidate integrations, storage and total costs; do not hide this tradeoff.

Initial mathematical identities are a preflight only. The subsequent isolated
Go implementation and independent ledger reference now exist; seven ordinary
unit roots pass. Full frozen race/replay, experiment and performance validation
are still pending at this dispatch version, not empirical success.
Test ALL40 consumed worlds and all schedules with same policy budgets,
uncertainty/random controls and original quality/recovery gates. No untouched
confirmation opening unless an explicitly frozen, justified follow-up passes.
Window/source-dependence remain separate leads if anchoring fails. Production,
sealed tasks, whitepaper and publication stay untouched; all seven goals OPEN.
