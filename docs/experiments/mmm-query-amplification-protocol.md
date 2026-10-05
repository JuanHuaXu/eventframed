# Bounded Bernoulli dependence amplification

Freeze this direction before scoring. Prior global/four-cell shrinkage chose
the strongest admissible endpoint; this motivates, but does not validate,
strengthening. No new case partitions or hyperparameter search.

For query marginal p and target conditionals f0,f1, let b=(1-p)f0+p*f1 and
d_y=f_y-b. Positive binary conditional probabilities require
lambda<(1-b)/d_y for d_y>0 and lambda<b/(-d_y) for d_y<0.
Let L be the minimum applicable bound, infinity for both d_y=0. Define
kappa=min(2,1+.99*(L-1)), with kappa=2 in the zero-motion case.
Use lambda=1+alpha*(kappa-1), alpha in[0,1]. Then the strengthened conditionals
are f_y+alpha*(kappa-1)*d_y. This is affine in alpha, retains both marginals,
and leaves positive distance from the boundary. The fixed factor2 limits
amplification; .99 avoids zero probability. These are frozen, not tuned.

Fit one global alpha by the existing convex conditional-log-loss optimizer,
phase0 actual query/target labels only, same equal-history/candidate/target
weights as dependence-v1. Compute the bound after target aging. No teacher,
case ID, phase1 label or future arrival enters fit features. Archived missing
query labels remain declared offline supervision. Data are consumed, not fresh.

Forecast diagnostic: compare independence, original(alpha0), fitted amplification
on all31 targets. Selection: evaluate calibrated Brier-information gains on
the original eight visible targets153..160; unchanged actual-publication C
forecasts score the selected candidate. No paid-query budget change. Selection
and modified-forecast tests are separate; pairwise validity does not instantiate
a multivariate process.

Retain2688 records/84cells, original random/entropy/joint controls and all failures.
Same four screens as local-dependence-v1: forecast joint log loss/target-expected
Brier versus independence+original, and unchanged-publication actual sampled/
teacher-weighted population Brier versus random+entropy. Nonharm lower>=-.001
every phase1cell; transition lower>0 on cases19,20. Use existing mean +/-3.5SE
descriptive trajectory bounds, no sequential confirmation claim.

Test strict validity, marginals, relabeling, zero motion, alpha0 reference,
convex fitting, independent derivatives, source/phase isolation, original control
replay and exact summary replay. Measure selector cost with precomputed laws
separately from posterior fits and serving. No production or whitepaper promotion.
