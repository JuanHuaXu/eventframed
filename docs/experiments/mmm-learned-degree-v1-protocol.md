# Learned degree weights v1, frozen protocol

Frozen before component/full-case evaluation, 2026-09-15. Compare learned
degree budgets with fixed noise1 versus learned noise, on all2688 consumed v120
records, each with32/64 most recent arrived labels and32-frame publications.
Retain linear64/32 controls in the six-arm output and every prior broad gate.
This is exploratory, not untouched confirmation. No production/paper changes.

For signed y=2Y-1 define G_d(x,x') as the mean signed-product kernel over all
binomial(9,d) features of degree d, d=1..4. Set
C = 11^T + sum_d t_d G_d(X,X) + v I.
The intercept variance remains1. Minimize F=(y^T C^-1 y+logdet C)/2, omitting
the parameter-independent Gaussian normalizing constant. The exact gradient
is (tr(C^-1 dC)-alpha^T dC alpha)/2, alpha=C^-1 y, including the log-parameter
chain rule. Inverse actions use Cholesky triangular solves, not a generic inverse.
The dense inverse needed for traces is assembled from unit-vector solves.

Initial degree totals t_d=9, v=1, matching the preceding equal-degree model.
Each t_d is constrained to[1e-4,16]; learned v to[.01,4], fixed v stays1.
No hyperprior, warm start, cross-window evidence multiplication or test tuning.
Use at most16 projected-gradient steps in log parameters; normalize the active
gradient by max(1,its infinity norm); projected trial bounds, Armijo1e-4,
8 backtracks starting at1 and halving. Projected-gradient infinity norm<=1e-6
stops early. At most145 objective evaluations including gradient evaluations.
On work-cap/line-search-cap, publish the best finite accepted iterate and record
the stop reason and remaining projected gradient. This is explicitly a bounded
approximation, NOT convergence certification; numerical invalidity is an error.
Record initial/final objectives, trace, parameters, evaluations and gradient for
every learned fit. A successful training objective is not predictive success.

Prediction is clip((1+k(x,X)alpha)/2,1e-12,1-1e-12), held fixed until next fit.
This empirical Gaussian working model for binary labels is not an ordinary
Bernoulli posterior and does not integrate hyperparameter uncertainty. All input
eligibility and no-future/no-missing-label gates remain unchanged.

Six arms: linear64/32, learned-orders/fixed-noise64/32, learned-orders-and-noise
64/32. Reuse840 non-harm and128 recovery gates per candidate against linear,
logistic,generic,Boolean,Markov exactly as spectral-regression-v1 (Boolean is not
a recovery control). Add paired comparisons against fixed equal-degree and
scalar priors, and fixed-versus-learned noise, with the same .01 non-harm and
.005 positive-lower-endpoint gain thresholds. Full traces and negatives retained.
No aggregate improvement substitutes for passing full requirements.

Validation: finite-difference all log-parameter gradients, Cholesky residuals,
monotone accepted traces, parameter/work bounds, initial-state agreement with
fixed-degree ridge, teacher/future/unavailable-label poison tests, deterministic
full replay. Independently reconstruct selected learned-fit objectives/gradients
and forecasts with another solver. Report cap rates and component timings;
neither proves loaded serving performance. Complexity O(J*N^3+J*5*N^2+N*256),
where J counts bounded objective/gradient evaluations, N<=64; predictions O(256).

Sources: [Duvenaud et al.2011, sections3.1-3.3](https://proceedings.neurips.cc/paper_files/paper/2011/file/4c5bde74a8f110656874902f07378009-Paper.pdf)
motivates learned order weights. [Rasmussen and Williams2006, chapter5, eq5.9](https://gaussianprocess.org/gpml/chapters/RW5.pdf)
supplies the Gaussian marginal-likelihood derivative and explicitly discusses
local optima. Our caps, projection, starting point and binary working-model
adaptation are declared research choices, not those sources' guarantees.
