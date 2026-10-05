# Actual mixture-predictive law by characteristic-function inversion

Historical attribution: Gil-Pelaez (1951), *Note on the inversion theorem*,
[DOI](https://doi.org/10.1093/biomet/38.3-4.481). The publisher PDF could not
be fetched in this session. The CDF inversion formula was checked in section
3, equation 21 of the authors' methodological paper
[Computing the aggregate loss distribution...](https://arxiv.org/html/1701.08299).
Its numerical quadrature is not automatically our error certificate.

## Derivation for this fitted mixture

From the conditional intercept, logit L=c+epsilon+sum_j a_j beta_j,
c=v0*ksum, a_j=x_j-v0*t_j, epsilon~N(0,v0) independent of the mixture
coefficients. Let m_j=a_j*slabMean_j and d_j=a_j^2*slabVariance_j. Then

phi_L(u) = exp(i*u*c-v0*u^2/2)
           * product_j [(1-g_j)+g_j*exp(i*u*m_j-d_j*u^2/2)].

This is the entire variational predictive logit distribution, not its
moment-matched Gaussian. For independent standard logistic Z,
E[sigmoid(L)]=P(L-Z>=0). Substitution Z=log(U/(1-U)), U uniform, gives
phi_Z(u)=Beta(1+i*u,1-i*u)=pi*u/sinh(pi*u). CDF inversion therefore gives

P = 1/2 + integral_0^infinity Im(phi_L(u))/sinh(pi*u) du.

At zero the integrand limit is E[L]/pi. Logistic convolution makes the CDF
continuous and the characteristic function integrable; all moments of this
finite Gaussian mixture are finite. The absolute tail above T is at most

integral_T^infinity 1/sinh(pi*u) du
 = (2/pi)*atanh(exp(-pi*T)).

The result is exact for the variational mixture before numerical integration,
not for the unknown true posterior. A zero mean alone does NOT imply P=1/2:
an asymmetric mixture can have zero mean. Tests must include that counterexample.

## Frozen numerical component contract

Use T=10 (analytic tail below 1.45e-14). Subdivide into
max(8,ceil(T*(1+M+sqrt(V))/pi)) initial intervals, where
M=abs(c)+sum abs(m_j), V=v0+sum d_j over nonzero-inclusion components.
This resolves potential phase oscillations and narrow Gaussian scales; it is
a conservative engineering resolution rule, not a quadrature theorem.
Use adaptive Simpson with minimum recursion depth 1, maximum depth 20,
total absolute error-estimate target 1e-10, and at most 65,537 integrand
evaluations. Charge every evaluation, including repeated interval endpoints.
Reject infeasible resolution, exhausted budget, nonfinite values and results
outside [0,1] by more than 1e-8. Report any smaller range clamp. Prediction
wrapper retains the declared 1e-12 probability floor separately.

Verify against point-logit sigmoid values, Gaussian integration, explicit
small-mixture enumeration, complement symmetry, zero-mean asymmetry, input
rejection and budget exhaustion. Preserve conditional-intercept transformations
when constructing a query law. Benchmark prediction separately from fitting.
No research-tape prior choice or quality result is supplied by this component.
