# Predictive validation after v92

Status: proposal only. Do not retune the consumed v92 confirmation prior.
The skeptical family still fails the complete gate, including a real sparse
majority tail with Brier0.40899. Prior adjustment is not enough evidence of safety.

## Primary-source rationale

Yao, Vehtari, Simpson and Gelman,
[Using stacking to average Bayesian predictive distributions](https://sites.stat.columbia.edu/gelman/research/published/stacking.pdf),
Sections2,3.2 and3.3, combine predictive distributions using cross-validated
proper scores rather than marginal-likelihood family probabilities. Their
leave-one-out construction motivates testing predictive mixture weights here.
The paper explicitly warns of unstable small-sample weights; its asymptotic
arguments do not establish our16-label non-harm condition. We have read these
sections, not merely adopted the abstract's recommendation.

## Proposed binary-score specialization

For eligible observations D, let g_i and p_i be generic and parity predictions
of Y_i=1 fitted without observation i. Set d_i=p_i-g_i. Our proposed two-family
Brier objective with a declared optional shrinkage lambda>=0 is

$$
\min_{0\le a\le1}\sum_i(g_i+a d_i-y_i)^2+\lambda a^2.
$$

Its one-dimensional solution is

$$
\hat a=\mathrm{clip}_{[0,1]}
\frac{\sum_i d_i(y_i-g_i)}{\sum_i d_i^2+\lambda}.
$$

Choose a=0 when denominator is zero. This is our explicit binary quadratic
specialization; lambda and its scaling must be frozen before the experiment.
The final predictive law mixes full-data generic and parity forecasts using
a. These are predictive optimization weights, NOT posterior probabilities of
the families. Do not relabel this operation ordinary Bayesian model averaging.

In a current iid component window all labels in D have already arrived. LOO
excludes the scored label from fitting but may use observations later in the
window; it is NOT an as-of historical forecast. A temporal/delayed extension
must use a valid forward or blocked scheme and test regime drift explicitly.

## Exact shortcut to verify, not assumed performance

Naively, LOO would refit each model n times. Conjugate counts permit a proposed
exact shortcut. For generic subset S, let c be the count in observation i's
assignment cell and s its positive count, both including i. Then

$$
q_{G,S,-i}(Y_i=1)=\frac{s-y_i+1/2}{c}.
$$

For parity S, let n be total count, u its agreement count, and b_i indicate that
y_i agrees with the parity rule. The leave-one-out agreement probability is

$$
\bar\theta_{S,-i}=\frac{u-b_i+1/2}{n}.
$$

Orient this probability to Y_i=1 using the rule evaluated at x_i. For each
family and rule, obtain leave-one-out marginal likelihood from

$$
\log L_{F,S}(D_{-i})=\log L_{F,S}(D)
-\log P_{F,S}(y_i\mid x_i,D_{-i}).
$$

Normalize subset weights using these LOO likelihoods and the unchanged subset
prior. All counts are positive where denominators are used. Proper Beta priors
keep event likelihoods nonzero. Compute in log space. This is an algebraic
proposal, not a claim that compiled n-fold refits were already benchmarked.

It avoids n separate full predictive compilations: after collecting counts,
LOO evaluation needs bounded subset work per observation, followed by one final
compilation. Verify exact agreement against explicit refits on small fixtures,
including constants, contradictory labels, permutations and impoverished support.

## Required falsification

- Freeze one unregularized baseline and any regularized candidate before fresh
  evidence. Do not search lambda on confirmation or remove the v92 tail case.
- Compare generic, skeptical BMA and stacking on identical fresh samples with
  unchanged component gain/non-harm gates. Keep n16 and all partial-view tests.
- Preserve the v92 tail as a consumed regression diagnostic, never as a new
  confirmation case. Report large per-fit harms as well as average gates.
- Learn weights only from training data, not the simulator oracle or final
  scoring domain. Prove that label removal also removes its likelihood weight,
  not merely its cell mean; otherwise LOO is contaminated.
- Account for fit memory, LOO work and adaptation delay. A component pass cannot
  authorize delayed-stream adoption or repair missing observation coverage.
