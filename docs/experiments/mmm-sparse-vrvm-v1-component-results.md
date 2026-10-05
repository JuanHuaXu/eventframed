# Sparse variational Gaussian update: component results

## Verdict

Gaussian-update component PASS. This is not yet a sparse variational learner
and supplies no new accuracy, calibration or whole-goal success result.
All seven research directions remain OPEN.

The research-only `_test.go` implementation solves in sample space using
Woodbury. It retains a low-rank covariance factor, coefficient means and
diagonal variances. It does not prune features or learn Gamma precisions.
The source and derivation are in the component protocol.

## Checks

Race tests compare six configurations (n=1,7,64 and p=10,256) against an
independent coefficient-space Cholesky implementation. They check196908
covariance entries, all means and stored diagonals, plus five logit-moment
queries per configuration. Precisions are nonuniform; repeated inputs with
conflicting labels are included. Input arrays remain unchanged.

A separate fixed-unit-precision10-feature control matches the existing
variationalStep. Its mean and variance agree on all512 inputs within1e-11.
Full-feature reference tolerances are1e-8 for means/covariance and1e-7 for
query variance; these are numerical test tolerances, not posterior guarantees.

Invalid dimensions, raw inputs, masks, duplicate masks, precisions, xi,
overflowing reciprocal precision and substantive negative variance are
rejected. Only roundoff-sized negative variance is rounded to zero. The mask
test was tightened so invalid-mask cases reach that check rather than failing
earlier on mismatched dimensions. No algorithm change was needed.

Artifacts: `mmm-sparse-vrvm-v1-component-tests.txt` and the tightened
`mmm-sparse-vrvm-v1-component-tests-verified.txt`. Both pass.

## Cost

Apple M4, Go darwin/arm64,64 samples/256 features, three300ms repeats:
847-858microseconds for one Gaussian update,139266bytes and1allocation.
This is not a full iterative fit or serving-latency measurement. Iteration,
Gamma/xi updates and predictive integration will add cost. The independent
dense reference is not included in the benchmarked implementation.

Work is O(n^2*p+n^3), storage O(n*p+n^2+p), bounded here by n<=64,p<=256.
The old quadrature helper assumes variance<=10 and cannot be reused blindly
with these per-feature precisions. No inference or performance guarantee
outside the verified numerical domain is inferred.

## Next

Implement and independently check Gamma/xi updates and the complete variational
bound, then prediction integration, as-of adapters, quality and full-fit cost.
`research/sparse-vrvm-elbo-derivation.md` records an algebraic cancellation
that avoids needing digamma for the specified Gamma factor shape; it also
warns not to mix old covariance with newly updated precision in logdet(S).
That proposed bound still needs executable verification before acceptance.
Freeze priors and stopping rules before inspecting quality outcomes.

Implementation:
`internal/observationlearners/sparse_vrvm_step_test.go`.
SHA256:
`c5de03b8e2cdd31f51fce0f32cfdf0417a7804c21e8b31850720508ed76a7efc`.
No production code, whitepaper, dependencies, commits or pushes changed.
