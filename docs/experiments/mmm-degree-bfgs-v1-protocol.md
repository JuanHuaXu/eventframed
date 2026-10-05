# Bounded quasi-Newton convergence screen v1

Frozen before evaluation, 2026-09-15. Same Gaussian objective, log-parameter
bounds, initial t_d=9/noise1, evidence origins and projected-gradient tolerance
1e-6 as learned-degree-v1. No predictive gate or parameter-prior changes.
New optimizer, not a claim that BFGS is itself new research.

Five dimensions permit a full-memory Hessian approximation and exact small
box-quadratic subproblem: enumerate lower/free/upper states for each unfrozen
coordinate (<=243), solve each free principal system, retain the feasible
minimum quadratic value, including the zero step. Positive-definite B starts
at identity. Line search along that feasible step uses Armijo1e-4 and at most
20 halvings. At most64 accepted steps (<=1345 objective/gradient evaluations).
Use Powell-damped Hessian BFGS: Bs=B*s, b=s^T B*s; if s^T y<.2*b, replace
y by theta*y+(1-theta)*Bs, theta=.8*b/(b-s^T y). Update B by
B-Bs*Bs^T/b+y*y^T/(s^T y). Reset identity if curvature quantities are tiny,
nonfinite, or the updated matrix fails Cholesky. Record resets and QP work.
This is our small-dimensional full-memory variant, NOT L-BFGS-B or its theorem.

Known interior/boundary quadratics, fixed coordinates, indefinite-matrix
rejection, free-system residuals and damped positive definiteness are component
tests. Data screen: all phases/cases/schedules, index0, publications0/128/224,
both windows and both noise choices:1008 fits. Use the same consumed subset as
the prior independent audit, selected without looking at gradients or outcomes.
Save parameters/objective traces/stop reasons and32 forecasts per fit. Compare
to the exact matching archived projected-gradient iterate, not a fresh-seed claim.

Convergence screen requirements: >=95% meet the same projected-gradient tolerance,
all final objectives <= old objective +1e-8, no failed numerical/eligibility checks,
exact replay. This is a prerequisite diagnostic, not any of the seven whole-goal
success criteria. Report predictive Brier changes even if optimization improves;
better likelihood must not be substituted for better prediction. A subsequent
full2688-run comparison must retain every original quality gate.

Source inspiration: [Byrd, Lu, Nocedal and Zhu, A Limited Memory Algorithm for
Bound Constrained Optimization](https://users.iems.northwestern.edu/~nocedal/PDFfiles/limited.pdf),
section2 uses gradient projection, quadratic models and feasible line searches.
Here exhaustive five-dimensional box solves replace its large-scale machinery;
we neither copy its implementation nor inherit its empirical claims. Damping
and positive-definiteness must be checked for this specific implementation.
No dependency installs, production edits, paper changes or pushes.
