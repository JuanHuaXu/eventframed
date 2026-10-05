# Within-pool ranking objective control

The prior critic predicts between-pool mean value but has negative within-pool
R-squared even in training. This motivates a distinct objective hypothesis,
not a proven root cause or permission to tune the failed critic on phase1.

Keep all ten features, phase0-only training, training-only normalization,
ridge penalty .01, tie rules, source artifacts, phase1 evaluation, and advancement
screens from the original critic protocol. Train the original critic unchanged
as the pool-value head. Train a separate rank head on per-pool centered features
and centered gain targets. Keep an intercept column of1 for the existing fitter;
center all other columns. Every pool has total training weight1.
Centered targets may span[-2,2]; halve them for the existing fitter's target
contract, then double rank predictions. This reversible scaling leaves the
linear ridge solution and penalty ratio unchanged; do not clip targets.

For inference in a pool, let c be the mean original-critic prediction. Apply
the rank head to the centered feature vectors; subtract their mean prediction
to get relative values r_i. Emit s_i=c+r_i. Thus mean_i s_i=c exactly up to
floating-point tolerance. This changes within-pool learning while preserving
the mean pool-value estimator. It does not calibrate c or the abstention rule.

Compare forced and zero-threshold gated modes with identical-cost random and
entropy controls. Retain the old critic as an unchanged baseline. All original
summary controls must match. Report both actual-answer and outcome-averaged
Brier across all84 cells; the same phase1 advancement screen uses actual-answer
Brier and cannot be relaxed. All data remain consumed, not fresh confirmation.

Tests: common within-pool feature/target shifts cannot change relative ranking;
preserved pool mean; pure-feature/target-field traps; detached replay; poisoned
phase1 targets leave both models and decisions unchanged; context head exactly
matches the old model; old control outputs and query counts match; byte-exact
summary replay. Measure component cost separately from Bayesian bundle creation.
No new feature/penalty sweeps, production, whitepaper, commits, or pushes.
