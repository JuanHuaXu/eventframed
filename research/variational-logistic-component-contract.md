# Frozen variational logistic component contract

Research-only batch approximation following Jaakkola and Jordan (1997),
Section2 and AppendixA, with the positive-curvature sign convention:
https://proceedings.mlr.press/r1/jaakkola97a/jaakkola97a.pdf

Every eligible window starts from N(0,I), ten signed features including the
intercept, xi=1 and no inherited sufficient statistics. Let
lambda(xi)=tanh(xi/2)/(4xi), limit1/8. Precision A=I+2sum lambda*x*x',
b=sum(y-1/2)x, mean=A^-1*b, covariance=A^-1. Update each
xi=sqrt(x'covariance*x+(x'mean)^2). Stop at maximum relative xi change
|next-old|/(1+next)<=1e-10, at most1024 iterations. Failure aborts, never silently
substitutes a MAP forecast. Supported window0..256, input0..511.

Evaluate the bound sum[-softplus(-xi)-xi/2+lambda*xi^2]
-logdet(A)/2+b'mean/2. Reject a decrease exceeding1e-10*(1+|previous bound|).
Return detached moments, bound, iteration count and residual; retain no samples.
The implementation currently repeats positive-definite factorization for each
covariance column: fit cost O(I*(N*d^2+d^4)), not an optimized O(d^3) matrix
implementation. The present loop also allocates one result per iteration.
These measured costs are explicit optimization opportunities, not hidden.

## Predictive law

Project to Gaussian logit mean x'mean and variance x'covariance*x. Numerically
average sigmoid over a positive normalized trapezoid measure with256 panels on
standard-normal z in[-10,10]. Reject nonfinite moments or variance outside[0,10]
except <=1e-10 numerical excess at the upper bound. The prior implies covariance
<=I, and signed feature norm squared10. The omitted normal mass is <1.6e-23;
discretization error and variational approximation error are separate.
Floor the forecast at1e-12 and1-1e-12, matching MAP.

Numerical QA compares1521 mean/variance pairs against an independent8192-panel
Simpson/tanh integral on[-12,12]; maximum observed error7.11e-15. This is not a
uniform error certificate, exact posterior, or calibration guarantee. All1152
consumed generator fits converge (maximum56 iterations), all589824 forecasts
are finite, and39366 uniform/nonuniform partial-law checks pass. Repeated-row
analytic covariance/mean checks include the harder256-label case, requiring571
iterations. Do not infer general quality from those numerical tests.

Full-input prediction costs O(d^2+257); optional compilation costs
O(2^9*(d^2+257)+9*3^9), with3^9 cells. Compiled query lookup uses a fixed
nine-coordinate index. Exponential tables are not a general high-dimensional
scaling strategy. Compilation is detached and conditioned on declared positive
input weights, with no query label or oracle access.

## Measured boundary

docs/experiments/mmm-variational-component-benchmarks.json preserves all33 rows,
source hashes, machine and command. On Apple M4,64-label fitting costs
.180-.201ms and compilation.667-.674ms; direct prediction1.200-1.215us,
compiled lookup7.152-7.252ns. The degenerate256-label fit costs27.535-27.757ms.
No other research computation ran during benchmarking. Excludes persistence,
queueing, serving concurrency, and LLM/backend calls. Not quality-matched yet.
