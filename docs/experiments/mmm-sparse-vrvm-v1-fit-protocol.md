# Sparse variational fit v1: frozen fitting contract

Before quality outcomes: use all256 degree<=4 Walsh features, including a
unit-precision intercept. Other coefficient precisions have Gamma(1e-6,1e-6)
shape/rate priors. No pruning, feature selection or covariance truncation.
The Gamma prior is proper but its marginal coefficient variance is infinite;
finite variational Gaussian moments are not a prior-variance guarantee.

Initialize each q(alpha) shape to A=.500001 and rate to A, hence expected
precision1; initialize xi_i=1. Every window starts fresh, with at most64
already-observed labels. No posterior reuse across overlapping windows.

Each iteration: exact Gaussian coordinate update using current expected
precisions and xi; Gamma rate B_j=1e-6+(m_j^2+S_jj)/2; xi_i=sqrt(E logit_i^2).
Evaluate the full bound on that explicit factor state after each coordinate.
Reject any nonfinite result or decrease beyond1e-8*max(1,abs(previous bound)).
Stop when consecutive full-iteration bounds differ by<=1e-6, or after64
iterations. Retain capped states as approximate, report all stops. The fixed
intercept has no Gamma factor. Final Gaussian and Gamma factors need not be
at a common fixed point at the cap; never silently refit one without costing it.

Use the compact bound in research/sparse-vrvm-elbo-derivation.md, verifying it
against the unsimplified Gaussian/Gamma prior and entropy terms. Validate
logdet(S) against dense covariance factorization and each coordinate's
nondecrease, including states before Gamma optimization. Retain Gaussian
moments, not a plug-in posterior-mean probability masquerading as integration.

Prediction integration and trajectory quality are subsequent required stages,
not established by this fitter. Existing variance<=10 quadrature is unsuitable.
No output law or success claim is issued until an independently tested
integration rule is frozen. Benchmark full fitting separately from one update.
