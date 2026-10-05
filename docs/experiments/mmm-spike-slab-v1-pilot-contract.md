# Spike-and-slab pilot: frozen before collection

This is exploration on the already-consumed v120 tape, not untouched
confirmation. Use all phases/cases/schedules, trajectory index 0, clock 128,
window 64: 84 fits and 2,688 subsequent forecasts. No outcome-based choice
of subset, prior, initialization, budget or prediction method is allowed.

Use all 255 nonconstant nine-bit Walsh masks of degree <=4 in ascending
mask order. Set independent inclusion prior pi=1/255, slab variance c2=1,
intercept variance 1. The prior expects one active coefficient; each query
has prior logit variance 1+255*pi*c2=2. This is a transparent sparse modeling
assumption, not a learned optimum or a guarantee of calibrated probabilities.
It replaces the original extremely diffuse Gamma prior. Do not approximate
this mixture by a Gaussian or tune pi after viewing these pilot outcomes.

Fitting follows the fixed-prior contract at cap 1,024; retain capped fits.
Predict using the actual mixture inversion contract, including conditional
intercept dependence. Save full factors, final xi, bound/motion traces,
query means/variances and quadrature accounting. Any fit/prediction error
stops collection explicitly; do not omit failing records or silently fallback.

Eligible evidence: initial 16 observations and earlier nonmissing observations
whose recorded arrival is <=128, retaining the latest 64. Compare origins
against the source publication. Only bits and admitted outcomes enter fitting;
the next 32 query bits enter prediction. Read evaluator Q/Y/control fields
only after producing forecasts. No teacher, case or seed information enters
the model. Fit and forecast are frozen throughout this publication window.

Report expected Brier, realized Brier, convergence counts and computational
cost; compare generic64, Boolean64 and Markov controls and logit-mean plug-in.
Markov is an existing system control with within-window updates, not a matched
cadence ablation. Report every phase/schedule cell, without selecting wins.
No one-index pilot confidence interval or full-goal validation claim.
Require byte-identical replay and future-label/evaluator poisoning tests.
