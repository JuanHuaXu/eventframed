# Frozen predecision query-value critic

Train one ten-feature ridge regression on phase0 delayed trajectories only,
with equal total weight per trajectory. Target is no-query expected Brier minus
the query's outcome-averaged expected Brier from the counterfactual diagnostic.
This is offline synthetic oracle supervision, not free real-world feedback.

Features, in order: intercept; (160-origin)/8; query Bernoulli variance times4;
query entropy/log2; original-probe gain times4; disjoint-probe gain times4;
original mean predictive Bernoulli variance times4; disjoint equivalent;
fraction of retained support with origin>=128; mean support age divided by176.
Only the two cached decision bundles may enter feature extraction. No source
Q/Y, future-arrival metadata, publication record, case/phase/index/seed, teacher,
or observed future loss may be read by feature extraction or selection.

Center/scale non-intercept features using phase0 weighted moments alone (use
scale1 if variance<=1e-12). Minimize mean weighted squared residual plus .01
times squared slope norm; intercept unpenalized. Use the existing Cholesky
solver. Verify normal-equation residual and an independent dense solver in tests.
No hyperparameter, feature, window or threshold search after seeing outcomes.

Choose the lowest-origin candidate within1e-10 of maximum predicted gain.
Forced mode always spends one query when the pool is nonempty. Abstaining mode
spends only if maximum predicted gain is positive; zero is not a calibrated
confidence threshold. Random-gated and entropy-gated controls use the exact same
trajectory mask, making their query counts match the abstaining critic. Also
retain no-query, ungated random/entropy and both old joint selectors.

Freeze model on phase0 before evaluating phase1, all21 cases,32indices,both
schedules. Both phases have been consumed before: this is a new frozen-model
comparison, NOT untouched confirmation. Report both outcome-averaged and actual
realized-answer expected Brier; the latter is the primary policy metric.
Report costs, abstention, and all phase/case cells with mean +/-3.5SE intervals.
These are descriptive, not simultaneous or anytime intervals.

An advancement screen requires every phase1 delayed cell's paired lower gain
bound>=-.001 versus both matched random/entropy controls, and positive lower
gain bounds for both switching cases19/20 versus both controls. Apply separately
to forced and gated modes; do not declare an overall success by mixing modes.
Passing this one-decision screen only motivates full-stream and fresh tests;
it cannot complete any of the seven goals.

Tests: independent solver; constant/degenerate data; strict oracle-field access
traps in feature extraction; training-only normalization; test-target poisoning
must leave fitted coefficients and decisions unchanged; deterministic replay;
all selected forecasts must resolve to the existing audited branch artifacts.
Hash source artifacts, model, scripts and protocol. Measure fit/inference cost
separately from producing the two Bayesian decision bundles (not included in
critic timing and not free). Preserve failures. No production/paper/push changes.
