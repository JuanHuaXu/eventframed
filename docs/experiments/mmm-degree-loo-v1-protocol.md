# Degree LOO v1 protocol

Frozen component and subsequent consumed-tape protocol. No production changes.
Source: Rasmussen and Williams, GPML chapter5, section5.4.2, equations5.12-5.13
<https://gaussianprocess.org/gpml/chapters/RW5.pdf>. Their inverse-diagonal
identity provides fixed-hyperparameter held-out means and derivatives. They
warn that squared-error fitting ignores predictive variance and leaves common
covariance scale unidentified. Here intercept variance and noise stay1; only
four degree variances vary. No uncertainty-calibration claim is inherited.

This is not v93's failed stacking of existing experts. It fits four kernel
degree variances within the fixed-noise model. It can still overfit a small,
changing window. Fixed-parameter leave-one-out predictions exclude the held
label, but hyperparameter optimization uses all validation labels. Only later
as-of forecasts are evaluation; the optimized LOO loss is not unbiased evidence.

Let signed y=2Y-1, C=11'+sum_d exp(u_d)G_d+I, V=C^-1, a=Vy,
d_i=V_ii, r_i=a_i/d_i. The fixed-parameter LOO mean is mu_i=y_i-r_i.
For D_j=partial C/partial u_j, a'_j=-VD_j a and d'_{ij}=-(VD_jV)_ii.
Then r'_{ij}=(a'_{ij}d_i-a_i d'_{ij})/d_i^2.

Two objectives, each32/64 labels, no extra tuning grid:
- Smooth surrogate: mean_i r_i^2/4, gradient mean_i r_i*r'_ij/2.
- Clipped Brier: mean_i(p_i-Y_i)^2, p_i=clip((1+mu_i)/2,1e-12,1-1e-12).
  Interior gradient is mean_i -(p_i-Y_i)*r'_ij; zero on saturated branches.
  Declare zero at clipping knots; no global smoothness/convergence theorem.

Start all degree variances9. Bounds[1e-4,16], noise1. Reuse bounded damped
BFGS with64 accepted steps,20 backtracks, Armijo1e-4, projected gradient1e-6.
Retain every stop reason, no replacing failures with tuned starts. At most1345
objective/gradient calls per fit; no warm-start evidence reuse. Forecast clips
the all-window posterior mean, identical output rule to the old degree model.
Working Gaussian regression on binary labels is not ordinary Bernoulli Bayes.

First verify explicit leave-one-out refits, held-label flips at fixed u,
finite-difference gradients, clipping branches, duplicates, singletons,
invalid inputs, and monotone finite optimizer output. Benchmark without other
jobs. Complexity O(4*n^3) per gradient, O(n^2) temporary storage, n<=64;
stored model remains256 coefficients. This is additional slow-path fit cost.

Then collect the original2688 v120 records, all publication clocks and paired
windows. Keep linear64/32 controls, raw surrogate64/32 and clipped64/32 as the
six output arms. Same as-of origins, no evaluator Q/teacher metadata, no aging.
Preserve existing840 non-harm and128 recovery gates per candidate and all five
comparators. Compare against old fixed-noise forecasts additionally. Full
replay, independent objective/forecast audit and model-level future/missing
poison checks required. Consumed evidence is not fresh confirmation; no whole
research goal is complete from component identities or optimization success.

Before the full collection, run a cost/convergence pilot retaining index0 from
each phase/case/schedule (84 records), every publication clock and all four
variants. This fixed pilot can expose solver/cost failure but cannot estimate
the32-trajectory quality intervals. Record its predictions and optimizer stops;
do not select a variant or retune starts from pilot outcome scores. Follow with
full collection only after numerical verification; preserve any capped fits.
