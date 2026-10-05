# Degree-prior v1: frozen comparison

Frozen before forecasts, 2026-09-15. All2688 consumed v120 records, unchanged
32/64 arrived-label caps, 32-frame publications, clipping and no aging. This
tests prior allocation, not hyperparameter optimization. No fresh confirmation
or production adoption. Preserve all spectral-regression-v1 negatives.

The previous full dictionary has coefficient variance1 for each of256 features:
degree-wise totals1,9,36,84,126. Distinguish a group-allocation hypothesis from
the simpler total-variance hypothesis with two fixed kernels, both diagonal37:

- Degree: intercept variance1; each degree-d feature variance9/binomial(9,d),
  d=1..4. Each nonconstant degree receives total variance9, preserving the old
  main-effect coefficient variance and all256 features.
- Scalar: intercept variance1; each of255 nonconstant features variance36/255.
  Same total nonconstant variance36, without allocating equal degree budgets.

These numbers follow feature counts, not fitted scores. For either kernel K,
alpha=(K(X,X)+I)^-1 y, y=2Y-1; prediction is clip((1+k(x,X)alpha)/2).
Equivalent weighted ridge penalty is sum beta_j^2 / variance_j. This is a
regularized signed-label regression, not an ordinary Bernoulli posterior.
All variances are strictly positive, so K is PSD and K+I is positive definite.
Models consume only eligible (X,Y), never evaluator Q or case identities.

Six output arms: linear64/32, degree64/32, scalar64/32. Reuse the independently
scored spectral-regression-v1 broad gates verbatim:840 non-harm and128 recovery
checks per candidate. Add paired degree-versus-scalar non-harm (all spans) and
terminal recovery-gain gates using the same thresholds. Also compare both against
archived unnormalized full ridge; report effects, not only aggregate pass counts.
Passing a subset, or outperforming a known weak control, is not overall success.
All original seven directions remain open pending their full requirements.

Component checks: explicit feature-sum and Hamming-kernel equality for all pairs,
diagonal37, solver residual, constant/duplicate/invalid inputs, exact unit-prior
equivalence to prior full ridge through an audit-only mode, as-of poison tests.
Independent solver checks all cases/schedules/phases at selected publications;
full deterministic replay required. Record fit and prediction component timing,
but do not claim loaded serving latency. No threshold/variance tuning afterward.

Source: [Duvenaud, Nickisch and Rasmussen, Additive Gaussian Processes (2011)](https://proceedings.neurips.cc/paper_files/paper/2011/file/4c5bde74a8f110656874902f07378009-Paper.pdf),
sections3 and3.1-3.3: sums/products of coordinate kernels, separate interaction-
order variance parameters, and symmetric-polynomial evaluation. Their work learns
kernel hyperparameters and evaluates regression tasks. Here the one-dimensional
kernel is the signed-bit product, orders stop at4, and variances are fixed by
the above counting rule. We do not inherit its empirical results or claim its
learned-kernel method has been tested. A positive result would still require
new-seed and selective/partial-observation validation.
