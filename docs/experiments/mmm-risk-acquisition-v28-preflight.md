# Target-risk acquisition v28: mathematical preflight

2026-10-02. V27's frozen overall rescue failed. This distinct research-only
candidate targets future-event prediction risk, not just global-hypothesis
information. At this preflight no new outcome cohort had been collected.
The subsequent [fresh study](mmm-risk-acquisition-v28-results.md) FAILS both
overall screens; the mathematical preflight is not an empirical validation.

## Exact Conditional Objective

Retain v27's declared theta/phi joint family and distinct training observations.
Let Y_e be the proposed training label and F_j a future target label. At the
current history let q_e=P(Y_e=1), q_j=P(F_j=1), and let nonnegative target
priorities v_j sum to1. The current Bayes Brier risk is sum v_j*q_j*(1-q_j).
The expected reduction after observing Y_e is exactly

$$
V(e)=\sum_j v_j\frac{\operatorname{Cov}(Y_e,F_j\mid\mathcal H)^2}
{q_e(1-q_e)}.
$$

This follows from total variance and the two conditional posterior means,
not from an assumption that the current model is the external truth. It is
a one-step expected utility, not a globally optimal multi-observation policy
or a guarantee of packet usefulness. Targets include future labels for both
observed and unobserved members. Already observed members cannot be queried
again in the current model.

For j different from e, the covariance is the posterior theta covariance of
p_theta(e) and the conditional future mean for j. For j=e additionally retain
E_theta[Var(phi_e|theta,history)]; for an unobserved member this is
E_theta[p_theta(e)*(1-p_theta(e))/3]. Dropping this term would again optimize
global learning while ignoring useful individual-rate learning.

## Bounded Shortcut and Falsifier

Build a weighted Gram matrix of the centered conditional future-mean vectors
over27 hypotheses. Each candidate's global covariance-square sum is a
quadratic form; restore the queried member's within-theta variance contribution.
This costs O(n*M^2) time and O(n*M+M^2) scratch rather than O(n^2*M) for direct
candidate-by-target enumeration. n<=200 and M=27 remain explicit caps.

The independent unit control enumerates positive/negative future predictions
after a real cloned-model Observe and compares the expected-risk difference
to every Gram score at multiple histories and frontier sizes. Negative
controls cover invalid priority vectors, state mutation and exhausted/seen
candidates. No hidden outcome, truth rate or undelivered label enters scoring.
The tests PASS for frontiers2,11 and150 at multiple histories, within2e-14
of independent positive/negative lookahead. Full module race tests pass. This
verifies the implementation of the declared model's objective, not its
empirical benefit. Fresh matched-cost and off-model cohorts are required
before any integration or success claim.

Three100ms Apple M4 benchmark repetitions measure130.07-130.72us per
150-event selection,35328 bytes and3 allocations. This is roughly16 times
the v27 information-selector component cost; it must be charged rather than
hidden behind equal label counts. These measurements exclude label acquisition,
storage, queues, model construction and full serving. No empirical outcome
benefit has been measured for this candidate yet.

Methodological source: [Roy and McCallum (2001)](https://groups.csail.mit.edu/rrg/papers/icml01.pdf)
motivates selecting observations by expected future error reduction. Their
Monte Carlo classifier method and empirical gains are not implemented or
inherited here; the covariance identity and Gram reduction are the exact
specialization to this finite hierarchical Bernoulli model.

All seven whole goals remain OPEN. Production and whitepaper untouched.
