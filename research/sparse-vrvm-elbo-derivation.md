# Variational interaction fit: next integration derivation

Provisional algebra for the next implementation; not a completed learner.
Use the Gaussian and Gamma coordinate structure from Bishop/Tipping UAI2000
section5, cited in the component protocol. This records our reduction of its
bound so implementation need not introduce a digamma approximation unnecessarily.

Let the intercept have fixed unit precision. For every other coefficient j,
the prior precision is Gamma(a,b), using shape/rate. Let
q(alpha_j)=Gamma(A,B_j), A=a+1/2, and q(w)=N(m,S).
Write v_j=S_jj+m_j^2, ell_i=phi_i'm and
h_i=phi_i'(S+mm')phi_i. Given xi_i>=0,

    L_likelihood = sum_i [log sigmoid(xi_i) - xi_i/2
      + lambda(xi_i)*xi_i^2 + (Y_i-.5)*ell_i - lambda(xi_i)*h_i].

Combining Gaussian conditional prior, Gamma prior, Gamma entropy and Gaussian
entropy cancels all log(2*pi) terms. Because A=a+1/2, the coefficient of
E[log alpha_j] is exactly zero. The remaining total bound is

    L = L_likelihood + p/2 + logdet(S)/2 - v_0/2
      + sum_{j>0} [a*log(b) - lgamma(a) - A*log(B_j) + lgamma(A)
                   + (B_j-b-v_j/2)*A/B_j].

This cancellation requires the specified Gamma shape; do not apply it to
arbitrary q(alpha). At the Gamma coordinate optimum B_j=b+v_j/2 the last
term vanishes, but retain it when comparing intermediate coordinate states.

The sample-space implementation supplies

    logdet(S) = sum_j log(D_j) - 2*sum_i log(L_ii),

where D is the inverse expected-precision matrix used for THAT Gaussian step
and L is its sample-space Cholesky factor. After updating Gamma rates, do not
recompute this determinant using the new precisions with the old covariance.
That would combine incompatible states and invalidate the monotonicity audit.

Proposed update order: Gaussian factor for current E[alpha], Gamma rates from
current v, then xi=sqrt(h), evaluate the complete bound on that explicit joint
factor state. Repeat from the new Gamma/xi state. At an iteration cap the
factors need not be at a common fixed point; label that approximation honestly.
Fit stopping, priors, initialization and caps still need a frozen protocol.

Before running quality experiments: verify this compact bound against a
separate unsimplified calculation and coordinate-wise nondecrease, test
logdet against dense factorization, validate prediction integration outside
the old variance<=10 domain, then exercise missing/future evidence isolation.
No retrospective LOO or simulator Q may select hyperparameters. The usual
full controls, proper-score evaluation and slow-path cost accounting remain.
