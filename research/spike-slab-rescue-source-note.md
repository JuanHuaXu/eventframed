# Rescue hypotheses after the multi-clock screen

Do not treat numerical convergence as correct inference, or assume that a
better fit necessarily improves future prediction. Separate fixed-prior
model mismatch, variational approximation error and publication cadence.

Primary research read: Xie, Zhu and Stephens (2026),
[A Flexible Empirical Bayes Approach to Generalized Linear Models](https://arxiv.org/html/2601.21217v1),
sections2.2,3 and appendixC.1. They optimize posterior means and prior
parameters via a second-order approximate objective, linked to normal-means
inference. Their logistic prior options include point-normal mixtures.
The Taylor objective is not automatically our Jaakkola lower bound; their
uncertainty caveat for correlated features still matters. Their prediction
target uses posterior means, so adopting it would not by itself preserve our
integrated predictive law. This is a possible replacement, not an implemented
or validated rescue. Do not reuse old bound certificates for that objective.

Secondary lead: [SuSiE authors' documentation](https://stephenslab.github.io/susieR/reference/susie.html)
describes structured single-effect regression and local-optimum refinement.
Its standard linear-regression implementation is not a logistic drop-in.
The original paper's publisher rendering could not be usefully searched;
do not claim its complete mathematics was verified here.

Next diagnostic, before any new prior tuning: compare the current conditional-
intercept variational approximation with direct two-dimensional quadrature
for one-feature Bernoulli logistic models under the SAME pi=1/255, unit slab
and unit intercept priors. Use deterministic balanced +/-1 designs at n16/64,
with noiseless and label-flipped cases. Include the null spike's separate
one-dimensional evidence integral and posterior predictive contribution.
Check two quadrature resolutions. Record exact-model evidence ratio, inclusion
and forecast against the approximate fit. No benchmark/Q/teacher data enters.

This is not an exact audit of all255 coefficients and cannot alone identify
the full screen's failure cause. If single-feature inference agrees closely,
do not sell another optimizer as the cure; investigate prior/candidate
competition, multi-effect posterior dependence and cadence separately.
