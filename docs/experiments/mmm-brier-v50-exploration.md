# V50 Brier-aligned aggregation lead

2026-10-04. Provisional alternative, researched while V49 normal cohorts run.
Do not change that frozen candidate, its cohorts or gates. No quality result
or adoption follows from this formulation. The suspected loss/objective mismatch
is a hypothesis, not a diagnosed cause of the hybrid's remaining failures.

[Vovk and Zhdanov (2009), Algorithm1 and Theorem1](https://jmlr.org/papers/v10/vovk09a.html)
give a strong aggregator for the categorical Brier game. For binary forecasts,
their two-coordinate loss is twice our single-coordinate squared loss.
Consequently use normalized weights w_k, advice p_k in[0,1], y in{0,1},

$$
g_y=-\log\sum_k w_k\exp\{-2(p_k-y)^2\},
\qquad q=\tfrac12+\tfrac14(g_0-g_1).
$$

After immediate observation, multiply each weight by exp{-2(p_k-y)^2} and
normalize. This is loss-based aggregation, NOT an ordinary Bayesian likelihood.
Retain original advice; adaptive experts may depend only on earlier observations.
Binary reduction gives q in[0,1]. The full algorithm's positive-part substitution
must also be checked independently at endpoints and finite precision.

Our explicit normalization/reduction derivation: setting two positive masses
to sum1 gives s=1+(g_0+g_1)/2 and q=(s-g_1)/2. The per-outcome domination
2(q-y)^2<=g_y telescopes with the mixture partition function. For an immediate,
static, full-support prior pi, single-coordinate cumulative loss obeys

$$
L_N(q)\le L_N(k)-\tfrac12\log\pi_k.
$$

A declared Markov prior over expert paths gives the corresponding path-prior
penalty by the same partition-function argument, only under immediate complete
feedback. This is our extension to prove/test, not a cited delayed theorem.
Prior-positive paths only; zero-prior paths have no finite comparison bound.

[Joulani et al. (2013)](https://proceedings.mlr.press/v28/joulani13.html)
analyze delayed-feedback transformations. Original-position as-of replay is NOT
automatically one of those transformations. No immediate regret theorem is
claimed for delayed/canceled/missing labels, selective nomination, resets,
population risk, priority metrics, packet usefulness or nested integration.

Next isolated primitive: independent positive-part substitution, exhaustive
binary label paths, singleton/endpoint/symmetry/invalid-input controls, static
regret arithmetic, owned finite state, zero-label prediction and declared limits.
Then fresh untouched broad world cohorts with all original quality/cost gates
and matched log-score aggregation controls. No post-design fitting, changed
old thresholds, or new primary claim from numerical algebra alone. Keep child
learners identical; never confuse a strategy's loss weights with evidence trust
or Anti-Pigeon split authority. All seven WHOLE goals remain OPEN.
