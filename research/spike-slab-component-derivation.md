# Finite spike-and-slab logistic component

Source: Carbonetto and Stephens (2012), *Scalable Variational Inference for
Bayesian Variable Selection in Regression, and Its Accuracy in Genetic
Association Studies*, Bayesian Analysis 7(1), 73-108,
[author PDF](https://stephenslab.uchicago.edu/assets/papers/Carbonetto2012.pdf).
The logistic appendix, equations 24-31, supplies the mixture-factor and
conditional-intercept structure. Our proper unit-variance intercept prior
changes the collapsed quantities below. This is not a full varbvs port, and
its empirical claims or exact-posterior guarantees do not transfer.

## Model and approximation

For nonconstant Walsh features X_ij in {-1,1}, y_i~Bernoulli(sigmoid(f_i)),
f_i=beta0+sum_j X_ij beta_j. Priors: beta0~N(0,1), and independently
beta_j~(1-pi) delta0 + pi N(0,c2), 0<pi<1, c2>0. Both pi and c2 are
inputs to this algebra component; no research-tape values are selected here.
The slab has finite variance. It is not the regularized horseshoe.

For fixed xi_i>=0, W_i=2 lambda(xi_i) and k_i=y_i-1/2. Define
v0=1/(1+sum W_i), t_j=sum W_i X_ij, ksum=sum k_i,
G=X' diag(W) X-v0 t t', b=X'k-v0*t*ksum.

Use q(beta_j)=(1-g_j)delta0+g_j N(m_j,s2_j). Let e_j=g_j m_j and
V_j=g_j s2_j+g_j(1-g_j)m_j^2. Preserve
q(beta0|beta)=N(v0*(ksum-t'beta),v0), not an independent intercept.

The profiled lower bound is

L = sum_i [log sigmoid(xi_i)-xi_i/2+lambda(xi_i)*xi_i^2]
    + (log(v0)+v0*ksum^2)/2
    + b'e - (e'Ge + sum_j G_jj V_j)/2
    - sum_j KL(Bern(g_j)||Bern(pi))
    - sum_j g_j/2 * [(s2_j+m_j^2)/c2 - 1 + log(c2/s2_j)].

Coordinate updates at fixed xi:
s2_j=1/(1/c2+G_jj),
m_j=s2_j*(b_j-sum_{h!=j}G_jh e_h),
logit(g_j)=logit(pi)+(log(s2_j/c2)+m_j^2/s2_j)/2.

For any query x, a_j=x_j-v0*t_j gives exact moments under this q:
mean=v0*ksum+sum a_j e_j,
variance=v0+sum a_j^2 V_j.
Update xi_i=sqrt(mean_i^2+variance_i). Reprofiling the conditional intercept
at the new xi can only increase the bound in exact arithmetic. This does not
make the approximation exact for the logistic posterior.

## Component checks and limits

Implement the collapsed bound independently of an uncollapsed joint-moment
calculation, including E[KL(q(beta0|beta)||p(beta0))]. Check agreement for
nonoptimal factor states, each coordinate's nondecrease, finite perturbations
around a coordinate optimum, label complement symmetry, invalid states, and
that inputs remain unchanged. Retain the intercept cross-covariance in both
paths. Check a one-feature update against numerical Gaussian integration of
its fixed-xi surrogate likelihood.

This component uses direct Gram construction for transparent auditing, not a
production implementation. A later efficient sweep can maintain X*E[beta]
instead of a dense Gram matrix and must match it. A future predictive layer
must integrate the spike-and-slab mixture, not silently substitute a Gaussian
with matching moments. Hyperpriors, fit convergence, predictive integration,
as-of pilot, quality tests and runtime integration remain unimplemented.
