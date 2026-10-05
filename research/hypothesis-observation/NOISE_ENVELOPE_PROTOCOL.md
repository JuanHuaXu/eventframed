# Frozen continuous-noise envelope rescue

Replace the four-vertex universal guard with a conditional-law envelope for
all common channel-noise values in [.10,.30] and all sixteen copied-renewal
masks. Preserve the .20 predictor, priors, local baseline, optimistic target,
epsilon=.01, all 900 quality gates, and five evaluation worlds. No risk-budget
or gain-threshold tuning. The interval is chosen from the already exposed stress
range: this is a research rescue on consumed cases, not fresh confirmation.

## Interval inclusion derivation

Set u=(noise-.10)/.20 in [0,1]. Every independent report likelihood is affine
and nonnegative in u (type2 is constant). Copies contribute either 0 or 1.
For a supported mask and latent hypothesis h, the joint history weight is a
product of these factors times prior1/16. Write it in degree-d Bernstein form:

    w_h(u)=sum_k c_hk B_k,d(u),  c_hk>=0.

Multiplying degree-(d-1) coefficients a by affine endpoints (l,r) gives

    c_k = (d-k)/d * a_k*l + k/d * a_(k-1)*r,

with absent terms zero. This follows directly by multiplying the binomial
basis terms by (1-u) and u. All coefficients remain nonnegative. With
m_k=sum_h c_hk and class law q_k(y)=sum_(h mod4=y)c_hk/m_k,

    q(y|history,u)=sum_k [m_k B_k,d(u)/sum_j m_j B_j,d(u)] q_k(y).

Thus every conditional law over the CONTINUOUS interval lies in the convex
hull of the normalized coefficient laws. Regret is affine in the target law,
so protecting those vertices protects every supported noise/mask law.
This is a directly derived finite polynomial envelope, not empirical coverage
of real-world noise. No oracle noise or mask is supplied to prediction.

Use the existing closest-feasible projection; do not discard solver failures.
The coefficient hull can be conservative: coefficient vertices need not
correspond to a single physical noise value. At most16*11 laws fit the
existing 256-law reference cap. Four observed types and ten reports only;
this is not a scalable unknown-source model.

## Checks and interpretation

Verify polynomial likelihoods and normalized conditional laws against direct
report products at endpoints and interior noise values; check malformed
history rejection. Independently recompute score identities, preserve all
original controls, and replay results. Expose any numerical convergence issue
rather than silently switching to a different projection.

A pass would still require new noise patterns, independent per-source rates,
adaptive acquisition and a justified real-data likelihood envelope. A failure
retains all thresholds and motivates population-risk allocation, not narrowing
the noise range after seeing results. No production or whitepaper promotion.

