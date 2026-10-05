# Provisional residual-logit calibration lead

Implemented and numerically verified; predictive quality FAILED the consumed
v120 diagnostic. See [results](../docs/experiments/mmm-logit-v120-results.md).
The proposal below is retained as the pre-result rationale, not a success claim.

The contextual-router pilot added essentially no
recovery gain over its global control. Previous global/time-only mixers also
failed. This lead changes the forecast combination family rather than trying
another prior over the same categorical expert choice.

## Source and distinction

[Heskes, Selecting Weighting Factors in Logarithmic Opinion Pools,
Sections2-3](https://proceedings.neurips.cc/paper_files/paper/1997/file/59f51fd6937412b7e56ded1ea2470c25-Paper.pdf)
describes normalized products of expert densities and the corresponding
logistic representation for binary classification. Its weighting assumptions
matter: unit-sum weights enter proofs, and nonnegativity aids interpretation
and control of overfitting. A nonnegative unit-sum logit pool still stays
between the minimum and maximum binary expert probabilities. It cannot remove
the oracle-hull limitation measured in our diagnostic.

The proposed residual calibration below is therefore an explicit extension,
not that source's constrained pool or an inherited KL guarantee. Nor is it the
failed v93 exact-LOO linear stacking algorithm.

## Declared candidate, before any results

Let u_i=logit(p_i) for the eight ORIGINAL issued expert forecasts, with i=0 the
generic64 baseline. Use features f=(1,u_0,u_1-u_0,...,u_7-u_0) and logit
`a_delta = u_0 + f dot delta`. Fit delta by minimizing

`sum_{j in H_t} [softplus(a_delta,j)-y_j*a_delta,j] + .5*||delta||^2`.

This is a convex penalized Bernoulli likelihood with a zero-centered unit
Gaussian coefficient prior, not a posterior-predictive uncertainty integral.
The forecast is sigmoid(a_delta), with the existing finite-score output floor.
Zero correction recovers the baseline within that shared floor. The intercept
and u_0 term permit calibration changes; the difference terms use disagreement.
Signed corrections can leave the expert probability hull, increasing both
possible headroom and risk. Correlated expert probabilities are covariates,
not independent extra observations of the outcome.

Use the existing32-frame publication cadence and paired64/32 eligible-label
caps. Every fit starts from the same prior; never feed an old posterior in as
a new prior while reusing its training window. Admission uses actual arrival
times. Training covariates are the forecasts issued at each label's origin,
not refitted forecasts made after its outcome. No teacher law, case identity
or hidden change time is permitted. Current outcomes follow prediction.

These are proposed fixed choices for an initial component test, not permission
to relax previous comparisons. Freeze a complete experiment protocol before
quality evaluation. Validate gradients, Hessians, solve residuals, prior-only
behavior, extreme probabilities and collinear forecasts independently. Then
test consumed-data headroom against baseline, Markov and retained variational/
segment controls across all21 cases. Any promising choice needs new seeds.

## Falsifiers and costs

Reject if calibration overfits stationary/noise cases, fails either switch
direction, or only improves after outcome-dependent case/parameter selection.
Compare simple baseline-only calibration with the full expert correction to
identify whether disagreement adds value. Report proper scores and accuracy;
improved ranking or confidence alone is insufficient.

For K features and N admitted labels, dense Newton fitting costs
O(I*(N*K^2+K^3)) with a properly reused factorization, not just O(N*K).
Cached prediction is O(K). This remains a slow-path candidate. No serving
performance or whole-direction success is implied by those operation counts.
