# Error evidence and functional form: frozen factorial

Compare four models: original ten features versus those ten plus the frozen
ten prequential features, crossed with linear versus full degree-two polynomial
regression. The polynomial basis is intercept, raw non-intercept features, then
all products x_i*x_j with i<=j. Dimensions are10/20 linear and55/210 quadratic.
This is a bounded interaction model, not a reproduction of LAL's random forest.

Use the centered-ranking architecture: context head on absolute gain targets,
rank head on within-pool centered expanded features and targets, preserving each
context head's pool mean. Reversibly halve centered targets and double predicted
relative gains. Standardize each expanded column using training-only moments.
Keep ridge penalty .01, equal weight per pool, origin tie tolerance1e-10, and
zero-threshold abstention. No feature, degree, penalty or threshold search.

Train each model on all phase0 delayed pools; freeze before evaluating all phase1
cases/indices/schedules. Use the same consumed targets, prequential projection
and decision artifacts. Retain actual-answer and outcome-averaged Brier, every
old query control, and identically gated random/entropy controls at matched cost.
Apply the unchanged advancement screen separately to each model and mode:
all21 delayed cells paired lower gain bounds>=-.001 versus both controls,
and positive lower bounds for both switching cases19/20 versus both controls.
Intervals remain descriptive mean +/-3.5SE over32 trajectories, not simultaneous
or anytime guarantees. Passing cannot complete a whole goal.

Base-linear must exactly reproduce the prior centered critic's decisions and
losses. Check polynomial construction, independent linear solver, target and
input ownership, mean preservation, phase1-target poisoning, normalization,
complete-delivery identity, hashes and byte-exact replay. Measure component fit
and decision times separately from Bayesian bundle/projection production.
No tuning against phase1, fresh-data claims, production, paper, commits or push.
