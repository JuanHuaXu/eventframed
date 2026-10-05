# Repeated-trial dispersion V34: isolated preflight

2026-10-03. V33's broad full-frontier rescue fails despite repairing much
curved-case calibration. This is research-only work on Goals1/4, not a
production correction or completion of another goal.

## Reasoning Gate

Confirmed: some affine Brier curves worsen beyond64 labels. Not confirmed:
which variance, mean-family, packet-selection or finite-sample effects
dominate. A plausible model mismatch is the fixed Beta2 member rate prior
when the actual generator has little between-member variance around its
mean function. Its falsifier is a fresh repeated-outcome comparison where
learnable dispersion fails to protect fixed2 while improving low-variance
cases. Do not retune the V33 data or declare a cause from curve shape alone.

For a declared mean p, phi~Beta(kappa*p,kappa*(1-p)), followed by conditionally
independent Bernoulli trials at phi, gives

    P(Y1=1 | p,kappa)=p,
    P(Y1=1,Y2=1 | p,kappa)=p^2+p*(1-p)/(kappa+1).

The first expression is independent of kappa. Distinct members with only
one observation cannot update an independent kappa prior; repeated trials
can. Changing fixed kappa after seeing V33 would be prior retuning, not
evidence that the one-label posterior learned dispersion.

## Joint Model and Boundaries

Retain V27's27 mean functions and priors .1/.8/25*.004. Independently use
uniform kappa prior over {.5,2,8,32,infinity}. Infinity is the degenerate
phi=p branch, not an improper Beta prior. Given h,kappa, each member draws
one persistent phi. All its trials share that phi; integrating it yields
the ordinary Beta-binomial likelihood. Kernel for the next trial is
(kappa*p+s)/(kappa+n), or p at infinity. Sequential pre-outcome likelihoods
and scored future laws use the same model. Adaptive, fixed2 and shared-only
controls differ only in their frozen dispersion prior support.

No fitted dispersion affects evidence authority, AP permissions, epochs,
rank deltas or production. This model is stationary and single-owner.
It neither generates hypotheses nor handles hidden changepoints. Trial
ordinals enforce next-trial ordering and reject replay/skips; callers
must establish actual independent trial/source identity. Copied answers
with new ordinals are NOT authenticated here.

Bounds:2..200 members,64 trials/member,135 states. Store O(N*27+135+N)
numbers and counts; predict/update O(135), independent of corpus size
but not advertised as full-serving constant-time. Log weights are retained
to avoid making numerical zero probabilities permanently absorbing.
Invalid ordinals/inputs/normalization leave prior state unchanged.

## Primary Sources and Adaptation

[Molenberghs et al. (2010), A Family of Generalized Linear Models for
Repeated Measures with Normal and Conjugate Random Effects](https://arxiv.org/abs/1101.0990)
distinguishes binary marginal variance from hierarchical/repeated-measure
dependence and discusses beta-binomial random effects. Our finite mean/
strength grid and replay contract are a new bounded adaptation, not their
combined normal-effects model or inherited empirical guarantee.

[Gelman (2006), Prior Distributions for Variance Parameters in Hierarchical
Models](https://sites.stat.columbia.edu/gelman/bayescomputation/Gelman2006.pdf)
motivates treating variance priors as substantive modeling choices. Its
normal-hierarchy/half-t recommendations do NOT directly prescribe our
discrete beta concentration grid. No claim of a universally good prior.

Before collecting fresh outcomes, audit the full integrated joint likelihood
and forecast against beta-integral enumeration, one-trial nonidentifiability,
multi-trial identifiability, exact fixed2 equivalence to V27 at one trial,
invalid/replay/ownership/cap controls, future-label flips and allocation/cost
limits. Freeze experiment/protocol/checker before outcome collection.
