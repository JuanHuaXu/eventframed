# Prequential ridge stacking V31 preflight

2026-10-03. This is a new isolated research candidate, not a repair to V30
data or gates. No V31 outcomes have been collected; all seven goals OPEN.
Production, existing tracked edits, and the whitepaper remain untouched.

## Audit Before Implementation

Confirmed: V30's actual weighted child law matches its recorded weights,
yet several consumed fixtures have significant true-rate mixture headroom.
Needs investigation: whether observable pre-outcome predictive performance
can exploit that headroom. Competing causes remain finite-label variance,
inadequate families, and an evidence-to-evaluation mismatch. An oracle does
not distinguish those causes or supply a deployable selector.

This research-only change is authorized by the ongoing experiment goal;
no upstream runtime fix, broad cleanup, deployment, or public release is
being inferred. Unit falsifiers are a nonoptimal/simplex-invalid QP,
recomputed post-label training rows, reused tickets, or hidden-label-dependent
decisions. A fresh failed outcome screen must stay failed regardless of
passing arithmetic, and cannot be rescued by changing this cohort's priors.

## Exact Declared Objective

Children are the same baseline, affine, and partition joint models as V30.
At issuance j retain their three Bernoulli predictive means x_j privately.
When its outcome y_j arrives, add that ORIGINAL x_j and y_j to the fit:

`w^+ = argmin_{w>=0, sum w=1} sum_j (w dot x_j-y_j)^2 + ||w-w_0||_2^2`,
where w_0=(.98,.01,.01). The ridge coefficient is1 over SUM losses; it is
not multiplied by the observation count. The new output is the convex
mixture of the updated child forecasts with these fitted predictive weights.
Weights are NOT posterior probabilities of the family index. No joint
Bayesian interpretation or automatic calibration certificate is claimed.

For a Bernoulli law, squared mean error against the binary outcome is the
strictly proper Brier loss (up to the conventional factor2). Hence a
convex combination of means specifies a whole Bernoulli predictive law,
not just a point summary. Proper scoring does not itself guarantee good
weights from finite, selectively observed, or dependent evidence.

The Gram matrix starts at I and the response at w_0. Each resolved ticket
adds x*x^T and x*y. Seven nonempty simplex faces exhaust constrained optima;
each face solves an at-most4x4 KKT system. Positive ridge makes the objective
strictly convex. Solver work is bounded independently of history length,
and no full history is replayed when feedback arrives.

## Ordering, Tickets, and Observation

Opaque owner/member/serial tickets bind immutable privately retained child
predictions. Other outcomes may arrive between issuance and resolution;
weight fitting uses the issued row, while child posteriors incorporate the
new label against their current joint state. Member identities and the
world are fixed during this bounded experiment. No epoch transfer,
adversarial source authentication, or changing member ontology is supplied.
Duplicates, foreign/replayed tickets, and duplicate pending nominations
reject before mutation. Unexpected child failure quarantines the parent.
Pending members are excluded from subsequent nomination but remain
predictable. Immediate Observe is an Issue/Resolve convenience, not a
license to recompute an already-issued delayed forecast.

Random and random-within-stratum nomination are unchanged in meaning.
Entropy uses the actual stacked law. An optional .2-exploration disagreement
policy uses the weighted variance of child means. That heuristic is not
Bayesian mutual information, falsification authority or an Anti-Pigeon
certificate. Record exact conditional probabilities and actual costs.

## Research Source and Limits

[Yao et al. (2018), Using Stacking to Average Bayesian Predictive Distributions](https://sites.stat.columbia.edu/gelman/research/published/stacking.pdf)
motivates fitting combinations using proper predictive scores rather than
family-truth probabilities. This bounded online ridge/prequential procedure
is our adaptation, not their batch LOO/PSIS algorithm. Their asymptotic
result is not transferred to this selective finite-label stream. Chronology,
held-out outcomes, stationary protection and computational cost still need
to be demonstrated under a newly frozen outcome protocol.
