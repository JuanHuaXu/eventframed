# Sparse variational predictive integration

Frozen before quality evaluation. Scored law will be Bernoulli(p), with
p=E_Z[sigmoid(mu+sqrt(v)Z)], Z standard normal, using the variational Gaussian
moments. This integrates q(w), not the exact posterior. No empirical calibration
or sparse-learning success claim follows from accurate integration.

For ordinary finite v, use adaptive Simpson integration over[-10,10], absolute
estimated tolerance1e-10, maximum16385 integrand evaluations and depth24.
Require minimum refinement depth3 before accepting an error estimate; a
component regression exposed coarse-tail false convergence without this.
Split at0 and the logistic transition -mu/sd and its +/-40/sd offsets when
inside the interval. Reuse evaluated endpoints. Exceeding either budget returns
an error, not an unnoticed approximation. The Simpson difference is an error
estimate, not a rigorous simultaneous certificate. Omitted normal mass is
erfc(10/sqrt(2))<1.6e-23; numerical discretization is separate.

When sd is huge, a controlled approximation avoids an unresolved step:
let L have standard logistic distribution. Then p=E_L[Phi((mu-L)/sd)].
Since Phi is1/sqrt(2*pi)-Lipschitz and E|L|=2log(2),
abs(p-Phi(mu/sd))<=2log(2)/(sd*sqrt(2*pi)). Use this branch only when that
bound is<=1e-10. This analytic smoothing bound does not bound library floating
point errors. At mu=0 symmetry gives exactly.5; at v=0 use sigmoid(mu).
Reject invalid moments. Apply the existing1e-12 probability floor only at the
final forecast boundary, separately from integration error.

Check against independent dense midpoint normal integration, symmetry,
zero-variance identity, old variance<=10 quadrature and extreme-scale cases.
Explicitly test work-limit exhaustion. Test fitted probabilities without using
future labels and benchmark integration separately from fitting. Full as-of
trajectory quality testing remains a subsequent stage.
