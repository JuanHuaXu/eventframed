# Sparse variational prediction: numerical results

## Status

Predictive integration and fitted-probability component tests PASS. The model
now exposes a Bernoulli forecast integrating its variational Gaussian, rather
than silently substituting sigmoid(mean). This is not empirical calibration,
exact posterior inference or a successful trajectory-quality experiment.
All seven goals remain OPEN. Fitting fixtures still cap at64 iterations.

## Confirmed Numerical Fix

The initial adaptive Simpson rule accepted a coarse tail integral too early.
At mean=-20,variance=.001, the estimate was1.53682975096e-9 and the symmetry
defect was5.12768e-10, despite estimated error7.52593e-11. This was an
integration defect, not a model or label-fitting change. Require refinement
depth3 before accepting the estimator. The original failed test log is kept.
The estimator is still not a uniform rigorous error certificate.

## Verification

-25 mean/variance pairs compare with independent131072-panel midpoint normal
integration. Maximum discrepancy1.25844e-11; maximum evaluations1940 on that
grid. Variances range.001 through10000, including the old<=10 domain.
-Symmetry, zero-mean and zero-variance identities pass, including extreme
finite moments. The old quadrature agrees within1e-8 where applicable.
-12 intermediate-large-variance cases (1e8 through1e19) satisfy the declared
logistic-smoothing bound around the normal-CDF limit. The very-large-variance
branch is tested separately; neither is falsely counted as midpoint accuracy.
-Explicit evaluation-budget exhaustion and invalid moments fail closed.
-A fitted16-label model returns finite interior probabilities on all512
inputs. This does not test prediction accuracy or future-label isolation in
a trajectory adapter that has not yet been implemented.
-The complete sparse fitting/component race suite reruns successfully,6.542s.

## Cost

Integral only, mean=.3,variance100:20.83-21.20microseconds,48bytes and1allocation
over three300ms repeats on Apple M4. Gaussian moment computation, fitting,
queueing and serving are excluded. Full fitting remains the earlier288-301ms
fixture; no sub100ms end-to-end claim follows from this integral benchmark.

## Next

Freeze a bounded as-of pilot with old controls and explicit model-fit costs,
then implement the adapter and collect outcomes. Do not use unrecorded teacher
metadata, widen evidence windows or retune priors. A cheap plug-in-mean
ablation may help isolate integration's contribution only if predeclared.
Any moment-cache optimization must be algebraically equivalent and verified
against current outputs. No quality improvement is claimed yet.

Artifacts: `mmm-sparse-vrvm-v1-predict-protocol.md`, original
`mmm-sparse-vrvm-v1-predict-tests.txt`, repaired `-predict-tests-verified.txt`,
extended `-predict-tests-final.txt`, and `-predict-benchmarks.txt`.
Implementation: `internal/observationlearners/sparse_vrvm_predict_test.go`.
SHA256:
`896be0e6ff096f1caa356def79a35f987987443943aa2dff7d10b1063fe4cf02`.
No production, paper, dependency, commit or push changes.
