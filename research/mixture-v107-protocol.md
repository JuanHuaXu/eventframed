# v107: consumed mixture headroom diagnosis

Frozen before diagnostic generation. This is not fresh confirmation, a rescue,
or a deployment gate. Use every delayed switch record in v106: two phases,
two directions, 32 trajectories each, publications 128/160/192/224. No filtering
by performance. Parent SHA256 is checked by the driver.

Regenerate initial samples and verify all stored X/Q/Y against the original
role-separated seeds. Reconstruct each long64/short32 model from the exact
stored as-of origin lists. Verify these lists from delay/missing metadata,
and compare reconstructed generic partial forecasts with issued predictions.
No post-publication labels enter fitting. Oracle targets enter diagnosis only.

Enumerate all 512 full inputs under the fixture's uniform input law. Store
four raw model predictions, the neutral forecast 0.5, and true Bernoulli q.
For each publication minimize mean expected binary Brier over constant convex
weights on (a) raw four and (b) raw four plus neutral. Also report all five pure
risks and the irreducible mean q(1-q). Weights are constant across inputs, but
selected with hindsight separately per publication. No oracle weight is served.

The objective is w'A w - 2B'w + C with A=mean pp', B=mean qp, C=mean q.
Enumerate nonempty simplex faces, solve their equality-constrained KKT systems,
and reject infeasible candidates. Near-singular faces may be skipped only when
the winning point has numerical convex first-order gap <=1e-8:
gap = grad(w)'w - min_j grad_j(w). Risk-gap is the convex lower bound in exact
arithmetic. Float64 checks are numerical diagnostics, not interval-arithmetic
proofs or statistical confidence bounds. No theorem for arbitrary non-PSD input.

A separate JS evaluator rebuilds quadratics from all prediction rows, checks
simplex feasibility, objective, gradient gap, pure-model/neutral nesting and
irreducible floor, verifies source hashes, and summarizes all 16 phase/case/clock
cells. Full Go reconstruction must replay exactly. Unit tests include singular,
interior, boundary and brute-grid cases. No external optimizer dependency.

Interpretation: low oracle risk suggests selection headroom within these models;
high raw-four risk relieved by neutrality suggests abstention/shrinkage headroom;
high five-way risk suggests a limitation of this constant-weight candidate class.
None proves that input-dependent weighting, different observation policies or
better learners are impossible. Full-input oracle risk is not the served
partial-observation Brier; do not subtract these as a policy improvement estimate.
No confidence interval or new 94.7% claim is warranted from this consumed audit.
