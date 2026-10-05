# Sparse variational interaction component v1

Source: Bishop and Tipping, UAI2000 (arXiv upload2013),
[Variational Relevance Vector Machines](https://arxiv.org/pdf/1301.3838),
section5, equations53-58. For logistic observations its variational Gaussian
has precision A+Phi'W Phi, mean S Phi'(Y-.5), W_ii=2 lambda(xi_i).
Gamma precision factors update from coefficient second moments; xi updates
from logit second moments. This is approximate inference, not exact Bayes or
a calibration guarantee. Their weak Gamma hyperprior does not imply a finite
marginal prior variance. Numerical sigmoid integration is a separate step.

Provisional design: multiple Walsh interactions, not a mixture of individual
parity atoms. Up to256 distinct degree<=4 masks over9 bits, including intercept.
Intercept can retain unit precision; other factors can use declared Gamma
priors. No hyperprior, iteration limit or pruning policy is frozen for the full
learner yet. First establish the shared Gaussian-update primitive and cost.

For fixed positive expected precisions A and xi, let D=A^-1,
U=W^(1/2)Phi, C=I+UDU', C=LL', F=L^-1 UD. Then
S=D-F'F, m=S b, b=Phi'(Y-.5).
Only F (n by p), D, m and diag(S) need be retained. Logit variance for phi
is phi'Dphi-||Fphi||^2. This sample-space derivation is our implementation
of the same Gaussian update, not a theorem about sparsity or learning efficacy.

Contracts:1<=n<=64,1<=p<=256, unique9-bit masks; positive finite precisions,
nonnegative finite xi, valid raw inputs; reject nonfinite arithmetic. Only
roundoff-sized negative variances may be rounded to zero. Check the Gaussian
normal-equation residual. O(n^2*p+n^3) work and O(n*p+n^2+p) storage;
no corpus-sized matrices. The offline dense reference is not the implementation.

Verify against existing10-feature fixed-prior variationalStep and independent
coefficient-space Cholesky at full256 features, including duplicates, conflicting
labels, nonuniform precisions, extreme inputs and rejected invalid values.
Race tests and isolated component benchmark required before integrating Gamma
updates, ELBO monitoring, bounded prediction quadrature or trajectory collectors.
Those remaining pieces must not be represented as already implemented.
