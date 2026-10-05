# Family blend V30: isolated research preflight

2026-10-03. No V30 outcome cohorts have been consumed. All seven whole
goals remain OPEN; production and the whitepaper are untouched.

## Pre-Patch Reasoning

V29 curved gains and linear losses on identical evidence confirm that
replacing the affine family with coarse partitions is not broadly safe.
Raw fixed32 traces also confirm zero odd-index coverage for deterministic
bit-reversal. These are separate model and nomination limitations. Other
credible causes include the prior's strong baseline mass and inadequate
32-label evidence; no new outcome results yet isolate those explanations.

This successor is a new isolated experiment, not a patch to V29 tapes,
thresholds, or serving. Its invariants are coherent model averaging and
outcome-independent random nomination within declared strata. A falsifier
is failure of the flat-joint likelihood enumeration or any duplicate,
unsupported-input, future-label, or selection-probability check. Even if
those pass, untouched linear/calibrated protection and nonlinear benefit
must pass before any proposed integration. No upstream production fix is
being adopted; this is a local research design.

## Joint Model

Let the family index have prior (.98 baseline, .01 affine, .01 partition).
Each child is a complete alternative joint evidence/future-outcome model.
The baseline uses member rates Beta(2b_i,2(1-b_i)); affine and partition
retain the frozen V27 and V29 hierarchies. One distinct training label per
member is allowed, and future labels share that member's latent rate.

For selected unseen member i and y in {0,1}, update
`w_m^+ proportional to w_m P_m(y_i=y | past labels)` and update the
child-conditional posteriors. The future law is `sum_m w_m P_m(Y_i=1|D)`.
Updating children is integration under alternative latent families, not
tripling the evidence. Cold law is .9999*b_i+.00005, not exact identity.
The baseline duplicated inside child families is part of this explicitly
declared prior, not an additional independent observation.

## Nomination

For n members use K=min(32,n) contiguous integer strata. Visit strata in
bit-reversed bucket order, but draw uniformly from their unseen members.
In the first K observations every member has positive marginal inclusion
probability 1/stratum-width. Conditional probability is zero outside the
currently scheduled stratum; this is not uniform per-step coverage.
After K labels, repeat the bucket schedule, skipping exhausted strata.

Also retain uniform-without-replacement and entropy policies. A family
information policy uses `H(sum w_m q_m)-sum w_m H(q_m)` with a .2 uniform
exploration floor. Its exact conditional probability is .2/unseen plus
.8 at the deterministic maximizer. The score separates family indices,
not every latent submodel, and is not an Anti-Pigeon divergence certificate.

The active design is a function of observed history, fixed coordinates and
independent policy RNG, not hidden current labels. Hence the action factors
cancel from ordinary joint posterior ratios; this is not a selected-stream
model for an unknown external observation process. No correction is claimed
for never-observed data, unknown source dependence, or misspecified models.

## Lifecycle and Cost Boundary

All state is single-owner and bounded to 2..200 members. Invalid or
duplicate evidence is rejected before any child update. Unexpected partial
child failure quarantines the parent and disables future predictions.
There is no concurrent publication, persistence, authority, or serving hook.
Prediction/update are O(55) bounded-state work per member; full-frontier
entropy/information selection is O(n*55), random selection O(n), and the
stratified scan is bounded by n across skipped buckets. Actual timing and
allocation measurements must accompany any modeled total-cost comparison.

## Research Sources

[Hoeting et al. (1999), corrected Bayesian Model Averaging tutorial](https://sites.stat.washington.edu/www/research/online/hoeting1999.pdf)
supplies the integrated-likelihood/model-weight framework, not a guarantee
of external predictive superiority for these new families.
[INL's Latin hypercube documentation](https://tmap8.inl.gov/source/samplers/LatinHypercubeSampler.html)
illustrates random draws within strata and cites McKay et al. (1979).
Our finite one-dimensional member design is not a full multivariate Latin
hypercube, and no variance or active-learning guarantee is inherited.
