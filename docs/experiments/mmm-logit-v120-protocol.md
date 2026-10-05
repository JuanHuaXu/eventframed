# Frozen consumed residual-logit comparison

Before quality outcomes: retain all2688 v120 runs,21 cases, both phases and
schedules. Four arms: baseline-only calibration64/32; eight-expert residual
calibration64/32. Every arm anchors to generic64, including the32-label arm,
so window length changes only meta-training evidence. The eight covariates
come from generic64/32, Boolean64/32, variational64/32 and segment64/32.

Use the [declared residual-logit model](../../research/residual-logit-lead.md):
offset u0=logit(p_generic64), features(1,u0) for baseline-only, and
(1,u0,u1-u0,...,u7-u0) for the full model. Input and output probabilities use
the existing1e-12 floor. Minimize SUM Bernoulli negative log likelihood plus
.5||delta||^2; all coefficients, including intercept, have zero-centered unit
Gaussian prior. This is a MAP/penalized-likelihood plug-in, not an integrated
posterior predictive or an inherited logarithmic-pool guarantee.

Fit every32 frames, using latest64/32 eligible origins. No initial16 meta
examples without original issued forecasts. A current outcome always follows
forecast; delayed/missing schedules retain actual availability. Each fit starts
from the same zero prior, not a recycled posterior on overlapping labels.
Training features use original issued expert laws, never refitted hindsight
probabilities. No teacher q, hidden rule, case identity or change time in fit.

Numerics:64 Newton iterations maximum,24 backtracks, Armijo1e-4, gradient
infinity norm<=1e-8. Stable objective differences follow the repo ridge fitter.
No solver failure can be dropped or replaced by a baseline fallback. Freeze
these settings after component QA; failures are reported, not optimized away.

Report expected Brier, accuracy and log loss for all256/terminal64 frames, plus
fit origins, coefficients, residuals, iteration counts and source hashes.
Require exact replay and1344 poisoned-prefix checks across four arms:
84 index0 runs *4 cutoffs *4 arms.

Criteria per baseline-only arm:336 non-harm cells versus generic64 and Markov12,
plus64 terminal recovery gains (eight changed cases, both phases/schedules,
same two controls). Per full arm:840 non-harm cells versus generic64, Markov12,
matched baseline-only calibration, matched variational and matched segment;
96 terminal gains versus generic64, Markov12 and matched baseline calibration.
Total2672 screens:2352 non-harm and320 gains. Keep all cases and controls.

Paired mean +/-3.5SE over32 trajectories; non-harm upper increase<=.01; gain
mean>=.005 and lower>0. These are approximate fixed-sample screens, not anytime
or simultaneous coverage. Consumed results cannot confirm a selected rescue;
fresh validation follows only if justified. No Go, production, private-data,
OpenClaw, whitepaper or publishing changes; no serving speed claim.
