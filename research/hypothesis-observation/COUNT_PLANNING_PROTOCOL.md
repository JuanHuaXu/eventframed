# Frozen exact finite count-state planning diagnostic

Before another rollout, compute the complete six-renewal Bayesian planning
optimum under the existing interval-aware joint model. Keep four initial
reports,16 total evidence credits, priors and outcome likelihoods unchanged.

For this model ONLY, renewals within each source are exchangeable: either
iid conditional on latent h/noise, or copies of that source's initial root.
The four initial outcomes plus eight renewal zero/one counts determine the
joint likelihood up to theta-independent action-selection factors. Do not
multiply by multinomial counts: we condition on an observed ordered history,
not a newly sampled unordered multiset. Child/parent likelihood ratios yield
next-report probabilities. Validate against ordered/permuted histories first.

For a fixed initial root pattern, counts sum<=6 gives C(14,8)=3003 states.
All16 initial patterns yield48048 states. Use memoized backward induction,
not receding approximation, with
  V(s,r)=min_a sum_y P(y|s,a)[Gini(s+a,y)+V(s+a,y,r-1)].
V(s,0)=0. This minimizes sum of SIX post-query model risks. Also evaluate the
original learning-area statistic (six PRE-query risks/6) and terminal risk;
they differ by the initial-versus-final boundary term. Do not claim optimality
for each separate metric.

Evaluate fixed, random, entropy, receding one-step, receding two-step and full
planning under the SAME joint model's prior predictive law. Average initial
patterns by their actual prior-predictive masses, which must sum to1.
These are exact finite model expectations, NOT fresh empirical confirmation
or the original84/108 gate screens. Preserve prior rollout failures.

Check count sufficiency, transition normalization, depth1/depth2 action costs,
Bellman inequalities at every state and telescoping of pre/post risk sums.
Replay the diagnostic and compare a separate forward occupancy evaluation.
No empirical accuracy, unknown-model robustness or serving performance claim.
No production changes or publication.

