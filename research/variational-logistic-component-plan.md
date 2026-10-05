# Variational logistic component: planned uncertainty comparison

Status: design only. The v118 MAP/context-tree results do not validate a broad
replacement. Ridge64 helps some additive tasks but fails its complete structural
target; ridge32 provides no required gain. Do not infer that Bayesian averaging
will repair this, or remove interaction-heavy protection cases.

This is the already proposed separate uncertainty experiment, not a ridge-penalty
sweep chosen from v118. Retain the same unit Gaussian prior, ten signed features
including intercept, eligible64/32 labels and publication cadence.

## Source-derived approximation

[Jaakkola and Jordan (1997), Section2 and AppendixA](https://proceedings.mlr.press/r1/jaakkola97a/jaakkola97a.pdf)
provide the quadratic likelihood bound and Gaussian variational updates. Using a
positive curvature convention:

lambda(xi) = tanh(xi/2)/(4xi), with lambda(0)=1/8;
A = I + 2 sum_i lambda(xi_i) x_i x_i^T;
b = sum_i (y_i-1/2)x_i;
Sigma = inverse(A), mu = Sigma b;
xi_i^2 = x_i^T Sigma x_i + (x_i^T mu)^2.

The evidence lower bound for this unit prior is
sum_i [log sigmoid(xi_i)-xi_i/2+lambda(xi_i)xi_i^2]
- log determinant(A)/2 + b^T mu/2.

The predictive integral uses the Gaussian projected mean x^T mu and variance
x^T Sigma x, rather than sigmoid(x^T mu). The paper's Section4 reports
underestimated posterior variance: this approximation is not an uncertainty
calibration guarantee or exact posterior.

## Required implementation contract

Rebuild each window from the original prior; do not reuse the previous posterior
and replay its same labels, which would count evidence twice. Freeze initial xi,
iteration/tolerance limits and convergence/failure behavior before quality data.
Keep all fitting in the research slow path. Use positive-definite factorizations;
check symmetry, solve residuals and finite covariance/probabilities.

Use a declared positive-weight quadrature for the one-dimensional predictive
integral, with weights normalized. Verify against an independent accurate
integrator across the supported mean/variance range. Agreement of two quadrature
orders alone is NOT a rigorous error certificate. If only numerical verification
is available, name the forecast a quadrature approximation and do not attach a
uniform approximation-error guarantee. Compile the resulting probabilities under
the same declared input measure and recheck all partial-law identities.

Test exact prior/no-information limits where defined, label complement and
feature permutation, rank deficiency, separable/constant labels, gradient or
fixed-point residuals, bound monotonicity within declared numerical tolerance,
and frozen-model ownership. Distinguish an evidence-bound increase from accuracy
improvement; the former cannot pass the latter's gate. Report nonconvergence,
approximation cost and any clipping explicitly.

## Fresh comparison, not adoption

Compare the approximate predictive against its same-prior MAP control, generic,
Boolean and context-tree controls at equal evidence. Preserve every v118 family,
both phases/schedules and fixed harm/gain thresholds. Use disjoint fresh seeds;
v118 can guide diagnosis but cannot be untouched confirmation.

This remains a limited additive model. Even a useful uncertainty correction needs
an independently tested incumbent-preserving composite; it does not make parity
linearly representable. A later admission mechanism must learn from actually
issued pre-outcome forecasts, not fitted training accuracy or teacher family IDs.
Neither this plan nor the component can close the seven-direction goal.
