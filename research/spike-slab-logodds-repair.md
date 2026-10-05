# Saturation repair audit

Confirmed: the unchanged pilot failed at finite log-odds36.745813783531418
because sigmoid rounded to1. A separate balanced64-observation synthetic
coordinate at xi10 reproduced failure at log-odds118.99846332987129.
Thus neither a corrupted tape nor nonfinite variance/mean explains the
failure. This is local untracked research code, not an upstream released bug.

Canonical factor state now stores finite log-odds, not a rounded inclusion
probability. For a=abs(odds), u=exp(-a), compute small=u/(1+u), large=1/(1+u)
and their log masses -a-log1p(u), -log1p(u); swap by sign. Entropy uses both
log masses; variance uses g*h rather than g*(1-g); predictive mixture factors
carry the exclusion atom separately. Never infer the tail by subtracting
a rounded dominant probability from1. Dense and matrix-free paths share the
scalar probability representation but retain independent profile/sweep algebra.

Finite-precision limitation: probabilities beyond the exponent range can still
underflow, although their finite log masses remain stored. Tests cover odds
through +/-1000 and explicitly distinguish representable tails (through700)
from underflow. This is not arbitrary clipping or arbitrary precision.
The original pilot prior, variational family, coordinate order, stopping
criterion, budgets and quadrature tolerances are unchanged. Floating-point
rounding differs, so old successful prefixes need not be byte identical.

Falsifiers: invalid/nonfinite odds accepted; lost representable exclusion
mass; disagreement with ordinary-probability KL in its safe range; dense/fast
disagreement; changed as-of evidence dependency; or the original pilot still
failing. Regression, adjacent-path and negative-control checks cover these.
Rebenchmark full fitting and prediction; no added I/O or production hot path.
All changed code is under internal/observationlearners/*_test.go.
