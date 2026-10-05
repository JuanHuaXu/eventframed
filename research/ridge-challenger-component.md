# Regularized logistic challenger: verified component, quality untested

The research-only challenger is implemented in
`internal/observationlearners/ridge_logistic.go`. Numerical tests and component
benchmarks pass. No fresh quality comparison, controller integration or deployment
has occurred. It does not replace the generic, Boolean or context-tree learners.

## Frozen mathematical and numerical contract

For x in the nine-bit input space, use z(x)=(1,2x0-1,...,2x8-1). All coordinates
are included without teacher relevance information. Minimize

F(beta) = sum_i [softplus(z_i dot beta) - y_i(z_i dot beta)] + ||beta||^2/2.

This is a summed likelihood with fixed ridge precision1 on all ten coefficients,
including the intercept. It is a penalized-likelihood/MAP plug-in under a unit
Gaussian prior, not an integrated posterior predictive. No uncertainty or
calibration guarantee is inferred from the optimum.

Start at zero. Use at most32 Newton steps, each with at most24 half-step
backtracks and Armijo coefficient1e-4. Solve the ten-dimensional positive-definite
Hessian by Cholesky without forming an inverse. Convergence requires computed
gradient infinity norm <=1e-8; exhausting a cap or failing a solve/search returns
an error and no usable model. No fallback silently claims a successful fit.
The ridge term gives strict convexity in exact arithmetic; this is not a formal
floating-point optimality certificate.

Training losses use stable softplus/sigmoid calculations without a probability
floor. Issued probabilities clip the fitted sigmoid to [1e-12,1-1e-12] as a
declared numerical forecast guard. This clipping is not evidence of calibration.
Optional compilation freezes those512 full-input probabilities under a supplied
positive input measure into all19683 partial states, matching existing interfaces.
The input measure is separate from the response fit; uniformity is not assumed
by the compiler. The model retains no samples and the compiled law is detached.

The implementation is motivated by the regularized-logistic literature reviewed
in [the proposal](soft-response-challenger-proposal.md). Its fixed penalty,
bounded Newton solver and intercept treatment are our declared implementation,
not a reproduction of glmnet's regularization path or variational Bayes.

## Confirmed numerical bug and repair

The first consumed QA sweep found a line-search failure on additive/stationary,
index21,N256 at QA base2176111600. At iteration6, gradient norm was9.73e-7,
above tolerance, but predicted improvement was only8.19e-14. Comparing rounded
total objectives near141.9475 reported the trial about1.4e-13 larger.

The repair computes the objective CHANGE directly. For small logit movement d,
softplus(a+d)-softplus(a)=log1p(sigmoid(a)*expm1(d)); orient a,d by the label.
The ridge penalty difference is beta dot delta + ||delta||^2/2. Compensated
summation reduces cancellation across contributions. Larger logit changes use
stable loss differences to avoid expm1 overflow. Armijo tests this difference;
neither the tolerance, iteration cap nor objective was relaxed.

The dedicated regression compares the small-step change against an independent
Simpson line integral of the finite-difference-checked gradient. A larger-step
comparison checks consistency with total-objective differences. The formerly
failing fit now converges under the original tolerance. This repairs numerical
evaluation, not prediction accuracy, and uses no new confirmation sample.

## Completed verification

Final `go test -race ./internal/observationlearners -run '^TestRidge' -count=1 -v`
passed (package1.467s). Scoped vet passed. Coverage includes:

- Independent objective expression, gradient and Hessian finite differences,
  Cholesky residual checks and rejection of singular matrices.
- Analytic scalar optima for constant labels and identical/rank-deficient inputs
  at sample sizes1/16/256, both labels and three input vectors.
- Label-complement and feature-permutation symmetry, deterministic repeated fits,
  no sample mutation and detached compiled-law ownership.
- All19683 partial states under each of uniform and nonuniform input measures:
 39366 independent completion-average checks.
- Input/size/iteration rejection, extreme logits, finite probability floors,
  invalid input mass, and explicit rejection of uninitialized/unconverged models.
-1152 consumed-generator fits:3 families*3 modes*32 indices*4 sizes(16/32/64/256).
  All converge; no accuracy/Brier result from these QA samples is a quality claim.

An initially weak ownership test changed a copy of the model; it was corrected
to mutate the original fitted coefficients and confirm the compiled law stays
unchanged. No existing production or historical experiment source was modified.

## Component cost

Apple M4,10 CPUs,16GiB RAM; Go benchmark suffix10. Three200ms repetitions per
configuration, with no other research computation running. All33 repetitions
and nine source hashes are retained in
[the raw benchmark artifact](../docs/experiments/mmm-ridge-component-benchmarks.json).
These use a deterministic consumed arithmetic fixture, not representative agent
traffic or a quality-matched production workload.

| Operation | Observed time range | Allocated bytes/op |
| --- | ---: | ---: |
| Ridge fit32 | 22.57-22.65 microseconds | 128 |
| Ridge fit64 | 35.67-36.02 microseconds | 128 |
| Existing subset fit64 | 6.240-6.307 milliseconds | about331776 |
| Existing Boolean fit64 | .421-.435 milliseconds | 16384 |
| Existing context-tree fit64 | 5.887-5.900 milliseconds | 116088 |
| Ridge table compilation only | 72.25-73.22 microseconds | about319488 |
| Direct ridge full-input prediction | 29.79-30.01 nanoseconds | 0 |
| Compiled mask63 lookup | 7.109-7.262 nanoseconds | 0 |

The fitted coefficient/statistics object occupies120 bytes (128-byte allocation).
The optional partial table occupies314928 bytes before allocation rounding.
Different learners have different representational capacity: cheap fitting does
not establish comparable quality. Raw subset/tree fits also construct full
prediction tables internally, whereas ridge coefficient fitting does not; the
ridge compilation cost is therefore reported separately. No service/queue/I/O,
concurrent persistence, p95/p99 or general throughput claim follows.

For N<=256,d=10,I<=32,L<=24, fitting costs
O(I*(N*d^2+d^3+L*N*d)), including objective-change evaluation. Working numerical
state is O(d^2); the returned model is O(d). Each trial evaluates a full objective
and its change separately, counted in the measured cost. Full prediction is O(d).
The nine-bit partial table is exponential in feature dimension, not a scalable
general solution: O(2^9*d+9*3^9) compilation and O(3^9) storage here.

Benchmark artifact SHA-256:
`d5ab57700598d9ace1df22347f683a2b97e45489f680b72a0cb2ae9366d882e2`.
All hashes and33 repeat rows were verified after recording.

## What remains

Freeze and run the equal-evidence full-input learner comparison on fresh transfer
data, retaining original Boolean-regime controls and all failed v116 outcomes.
Then evaluate a budget-matched composite under adaptive partial observation and
delayed feedback. The variational Bayesian extension remains proposed, not
implemented. No roadmap direction is complete from component correctness or
speed alone; no publication or production change is authorized by these tests.
