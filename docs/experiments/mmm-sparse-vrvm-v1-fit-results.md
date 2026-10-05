# Sparse variational fitting: bound and coordinate tests

## Status

Coordinate-update and bound checks PASS. No predictive-quality experiment or
calibrated output law is complete. All three fitting fixtures reached the64
iteration cap; do not label them converged. All seven goals remain OPEN.

The frozen fit protocol declares Gamma shape/rate priors1e-6, an independent
unit-precision intercept, all256 degree<=4 features, fresh per-window starts,
64 iterations and the complete variational-bound checks. There is no pruning
or tuning against simulator outcomes. The Gamma prior's marginal coefficient
variance is infinite; the approximation's finite moments do not change that.

## Verification

The compact bound agrees with a separate unsimplified calculation of Gaussian
and Gamma priors/entropies within1e-8, both before and after Gamma rate updates.
This explicitly tests the cancellation of the expected-log-precision terms,
not just the rate-optimal special case. The independent expression uses a
test-only numerical log-Gamma derivative; the implementation needs no digamma.
Sample-space logdet(S) agrees with dense covariance Cholesky at10 and256
features within1e-8. The stored determinant belongs to the Gaussian state;
subsequent Gamma updates do not rewrite it.

All Gaussian, Gamma and xi coordinate transitions pass nondecrease tests.
Final rates and xi match the stored Gaussian second moments. Fitting does not
mutate input labels and all512 raw-input moments remain finite on the fixtures.
Earlier dense-reference, original10-feature and invalid-input tests rerun under
race and PASS. Combined test package time6.238s.

| Labels | Iterations | Stop | Initial bound | Final bound |
| --- | ---: | --- | ---: | ---: |
| 1 | 64 | cap | -3163.34399 | -3162.58250 |
| 16 | 64 | cap | -3193.53382 | -3176.07383 |
| 64 | 64 | cap | -3237.91126 | -3218.40743 |

These are fitting bounds, not Brier scores. Their large offsets include the
declared weak-hyperprior terms. No quality conclusion follows from improvement.

## Cost And Remaining Work

Full64-label fitting fixture:287.6-301.2ms over three300ms benchmark repeats
on Apple M4, approximately9.05MB allocated and142-145allocations. This includes
all64 coordinate iterations and bound checks, unlike the earlier0.85ms single
Gaussian update. It is slow-path work, not serving-latency evidence. Precision
learning and repeated moment evaluation are not free.

The bound currently recomputes the same Gaussian training moments for its
Gamma and xi checks; caching them per Gaussian state is a possible exact-cost
improvement, not a measured speedup. Avoid broad optimization before obtaining
a scored predictive law and a bounded pilot. Preserve the nondecrease checks.

Next freeze and verify numerical sigmoid-Gaussian integration for variances
beyond the old<=10 domain. Then implement the as-of adapter, a cost/convergence
pilot and full quality comparisons when justified. Current components have
no scored forecast output and cannot support efficacy claims. No production
code, paper, dependency, commit or push changes.

Artifacts: `mmm-sparse-vrvm-v1-fit-tests.txt`,
`mmm-sparse-vrvm-v1-fit-benchmarks.txt`, frozen fit protocol,
`internal/observationlearners/sparse_vrvm_fit_test.go`, and the extended
Gaussian component's stored logdet. The earlier component source hash is
historical: this turn adds logdet storage and reruns all its tests.
