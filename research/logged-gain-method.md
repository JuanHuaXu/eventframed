# Independent logged usefulness: method and limits

## Research basis

Waudby-Smith, Wu, Ramdas, Karampatziakis and Mineiro,
[Anytime-valid off-policy inference for contextual bandits](https://arxiv.org/html/2210.10768v3),
sections1.1,2.1,3 and3.1, supplies the off-policy/DR framework and explains
predictable logging, positivity, varying policy values and history-dependent
interpretation. Their tighter empirical-Bernstein construction is not implemented
here. Karampatziakis, Mineiro and Ramdas,
[Off-Policy Confidence Sequences](https://proceedings.mlr.press/v139/karampatziakis21a.html),
is earlier work on time-uniform off-policy evaluation. Neither source establishes
EventFrame's effectiveness or supplies missing labels for it.

## Our bounded reference

Before action A, declare logging p, candidate c, control b and loss regression m.
All are distributions/vectors on a finite action set, with m in[0,1]. Require
positive p wherever either target policy has positive mass. Let d=b-c and
observe only the selected action loss L in[0,1]. Define

    X = sum_a d_a m_a + d_A / p_A * (L - m_A).

Conditional on pre-action history/context, averaging over the randomized action
and its outcome cancels m exactly, giving E[X]=sum_a d_a E[L(a)]. With known p,
the regression may be wrong without biasing that conditional mean. This does
not allow wrong propensities, unmeasured selection, outcome-fitted regressions,
missing feedback silently treated as zero, or unobserved actions assigned mass.

Compute predictable endpoints l,u by enumerating all logging-supported actions
and both loss endpoints0,1. Let R=u-l, S_n=sum X_i and V_n=sum R_i^2. For any
fixed positive lambda, conditional Hoeffding gives a nonnegative supermartingale

    exp(lambda * sum_i (X_i - E[X_i | past/context]) - lambda^2 V_n / 8).

Mix the eight frozen rates equally. Invert the mixture at2/alpha to obtain
b(V). The interval [S_n/n-b(V_n)/n, S_n/n+b(V_n)/n], intersected with[-1,1],
covers the running average conditional gain simultaneously in n by Ville and a
two-sided union bound. Predictable random ranges are allowed in this exponential
construction. Do not substitute them into an unjustified fixed-n random-radius
formula. Do not intersect intervals across n: the target average can change.
Zero accumulated range means all increments are conditionally deterministic.
Numerically nonzero tiny ranges are rounded upward, never down to certainty.

This conservative reference is intentionally distinct from variance-adaptive
methods. It can be too wide to make any useful decision; that is an empirical
question, not permission to relax coverage after seeing the outcomes.

## Verification and boundaries

Tests enumerate576 unbiasedness identities,4608 conditional MGF inequalities,
and4096 leaves of an adaptive six-step action/outcome tree. They check ownership,
positivity, duplicate/out-of-order evidence and atomic failed updates. A stress
test exposed overflow from an extreme user betting rate, producing false
precision; supported rates are now restricted to[1e-6,2] and this regression
is tested. The default grid is unchanged. These finite checks support the
implementation but do not replace the argument above or prove policy efficacy.

The caller must supply authentic unique outcomes, pre-action propensities,
predictable regressions/ranges and the declared action semantics. The API cannot
prove those facts from numbers. Use the frozen logged-gain-v1 protocol for an
isolated consumed-data audit before considering tighter estimators or promotion.
