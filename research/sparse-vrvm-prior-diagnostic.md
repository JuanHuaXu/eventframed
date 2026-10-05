# Prior-only diagnostic after the convergence screen

The 1,024-step screen meets the numerical stop on 57/84 fits but still trails
the original controls. Optimizer budget alone is therefore not a sufficient
explanation. Remaining hypotheses include prior misspecification, variational
approximation error, feature-family mismatch, and changing data. None is yet
proved to be the sole root cause.

## Exact implication of the existing prior

For precision alpha~Gamma(a,b), shape/rate, and w|alpha~N(0,1/alpha), integration
gives the Student distribution with degrees of freedom 2a and scale sqrt(b/a):

f(w) = Gamma(a+1/2)/(Gamma(a)*sqrt(2*pi*b))
       * (1+w*w/(2*b))^(-a-1/2).

For any M>0, substitute w=sqrt(2b)*sinh(u). Then

P(|w|<=M) = 2*Gamma(a+1/2)/(Gamma(a)*sqrt(pi))
            * integral_0^asinh(M/sqrt(2b)) cosh(u)^(-2a) du.

Because the integrand is <=1, replacing it by 1 is a rigorous upper bound.
This is a prior-only identity, unrelated to the evaluation tape or its labels.
For the existing a=b=1e-6, useful finite logit-coefficient ranges can have
very little prior mass. Compute the exact central probabilities numerically
and compare them with this upper bound before selecting any replacement.
Do not confuse finite E[alpha]=1 with finite E[1/alpha]; the latter diverges.

Verified by `TestSparsePriorCentralMass`: the substitution matches the exact
standard-Cauchy CDF when a=b=1/2. For a=b=1e-6, 4,096/65,536-panel integration
agrees within 1e-12 and stays below the analytic upper bound:

| Radius M | P(abs(w)<=M) | Analytic upper bound |
|---:|---:|---:|
| 1 | 0.0000145085518435 | 0.0000145086386253 |
| 5 | 0.0000177273748267 | 0.0000177275090279 |
| 10 | 0.0000191136436216 | 0.0000191138014372 |

Thus only about 0.00145% of this coefficient prior lies between -1 and 1.
This proves that it is extremely diffuse on the standardized coefficient
scale, not that it is neutral. It does NOT prove the posterior shares that
mass distribution or that this prior is the sole cause of the observed errors.
The test passed in 0.02 seconds (0.382 seconds package time).

## Source-grounded replacement lead, not an adopted rescue

Piironen and Vehtari (2017), *Sparsity information and regularization in the
horseshoe and other shrinkage priors*, EJS 11(2), 5018-5051:
[primary manuscript](https://arxiv.org/html/1707.01694), sections 2.3 and 3.5,
[published version](https://doi.org/10.1214/17-EJS1337SI).
Their regularized horseshoe uses conditional variance
tau^2*c^2*lambda_j^2/(c^2+tau^2*lambda_j^2), bounded above by c^2.
It separates sparsity from regularization of large coefficients. Their
non-Gaussian sparsity calibration is approximate, not an exact logistic
posterior theorem. A new prior would require a new inference derivation;
the existing conjugate Gamma update cannot simply be reused.

Possible next action: prior-predictive checks and a justified finite-slab or
global-local prior, with frozen assumptions unrelated to teacher identities.
This must remain distinct from prior failed scalar/degree shrinkage studies
and from mere post-hoc clipping of predicted probabilities. Do not pick a
prior scale or expected sparsity from the consumed test's best outcome.
The diagnostic below can establish a prior property, not that changing it
will improve real-world forecasts.
