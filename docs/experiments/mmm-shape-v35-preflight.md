# Member-rate shape V35: isolated preflight

2026-10-03. Previous goal turn was progress: V33/V34 completed independent
fresh studies. V34 confirms low-variance pooling gains but packed confidence
in the two-rate worlds remains about .90 for .80 usefulness. All seven
whole goals OPEN; production, whitepaper and earlier evidence untouched.

## Reasoning Gate

Confirmed symptom: packed signed bias remains near .10 after16 trials/member.
Candidate causes: member-distribution shape mismatch, selection conditionality,
mean-function misspecification and finite member evidence. V34's learned
Beta2 variance and persistent bias make shape a reasonable next experiment;
they do not establish it as the sole cause or identify a runtime bug.

Change the declared prior family, not recorded evidence or clipping gates.
Retain all V34 mean functions and ordinary Beta kernels. Add mean-preserving
two-point rate laws. Preserve fixed2/adaptive/local controls and test fresh
on/off-grid and continuous/unimodal worlds. A falsifier is failure to repair
bias/protect those controls under the frozen fresh evaluation. This is not
an upstream patch, data repair, or justified production change.

## Model

For each original mean p and width d=.05,.10,...,.45:

    lo=max(.01,p-d), hi=min(.99,p+d),
    mass_hi=(p-lo)/(hi-lo), mass_lo=1-mass_hi,
    phi in {lo,hi} with those masses.

Then E(phi)=p exactly, including clipped endpoints. A member draws one
persistent phi; trials are independent conditional on phi. Ordered member
likelihood is mass_lo*lo^s*(1-lo)^f+mass_hi*hi^s*(1-hi)^f. Its next-trial
law is the likelihood ratio M(s+1,f)/M(s,f). Cached conditional means
implement the same law through posterior atom odds, avoiding global refits.

Original five Beta/point kernels receive half the shape prior, retaining
their1/5 relative weights. Nine atom widths share the other half. Mean
function priors stay .1/.8/25*.004;27*14=378 joint states. This explicitly
changes the shape prior; any early improvement is not learned shape until
repeated trials provide shape-dependent likelihoods. One trial/member
must leave shape marginal weights at the independent prior.

This finite catalogue includes known problematic shapes but is not fitted
on the new study outcomes. Uniform width spacing is frozen before fresh
collection. Off-grid/asymmetric/triple-rate generators test whether benefit
extends beyond a specially matched two-point fixture. Posterior mass does
not identify real ontological groups, provide AP authority, or authenticate
independent sources. Trial ordinals prevent replay only.

Bounds remain2..200 members and64 trials/member. Predict O(378) dot product;
update O(378), including atom-odds arithmetic for one member. Cache storage
O(N*378), not constant in frontier size. Finite log weights survive numerical
underflow. Invalid/replayed/capped input or normalization does not commit.
No asynchronous/temporal/delayed integration is provided by this candidate.

## Primary Research and Limits

[Cai, Campbell & Broderick (2021), Finite Mixture Models Do Not Reliably
Learn the Number of Components](https://proceedings.mlr.press/v139/cai21a.html)
defines Bayesian mixtures and shows component-count inference can fail
under misspecification. We use it as a warning about interpreting mixture
weights as structural truth, not as an inherited theorem for our fixed
catalogue, covariate-dependent repeated-trial model. We do not infer an
unknown number of groups or claim nonparametric consistency.

The mean-preserving discrete catalogue, kernel cache and bounded tests are
our research adaptation. An exact finite posterior only proves consistency
with its declared model, not model correctness. Batch Beta integrals and
direct two-atom marginalization, cache/nonmutation/input/underflow controls,
future-trial flips, fresh cohort replays and measured cost precede adoption.

## Pre-Collection Verification

Batch integrated joint weights/laws agree within3e-10 at2,11,150,200
members through64 trials/member. Clipped two-atom means match p within
2e-15. First150 distinct trials leave all14 shape weights at their prior;
second trials change them. Unrelated member caches, input ownership,
prediction nonmutation, invalid/replayed/capped evidence, atomic invalid
normalization and supported-state underflow recovery pass. Future-label
flips preserve legacy/candidate prefixes on six nonlinear/novel regimes.

Apple M4 Go1.27.1 component benchmark:150 predictions13.907-14.727us,
16th-trial update9.650-9.687us, both zero allocations; construction
66.393-66.812us/500993-500994 bytes/six allocations. Update includes
restoring benchmark state. These are not serving or acquisition costs.
The frozen50ms phase-work/1MiB construction caps are declared separately
from V34's unchanged25ms cost rule. No V35 outcome cohort existed when
these model/benchmark checks were recorded.
