# Finite posterior calibration diagnostic

Frozen before quality scoring. Consumed v120 data, not untouched validation.
This is an invented finite conditional-model experiment, not a claim to recover
the continuous Gaussian posterior or the full nine-coefficient expert correction.

Declare 25 models: intercept a and slope deviation b independently on
{-2,-1,0,1,2}, probability sigmoid(a+(1+b)*logit(p_generic64)). Prior is a
half-mass identity atom plus half a normalized product exp(-(a^2+b^2)/2) slab.
Merge coincident identity atoms. Prior predictive is NOT the identity forecast;
include it as an explicit no-learning control. No tuning of this prior/grid
after inspecting outcomes. Input/output floor 1e-12, shared with earlier studies.

Compute exact finite posterior weights from Bernoulli likelihoods of ORIGINAL
issued forecasts and arrived labels. Predictions integrate those weights.
No teacher probabilities in fitting. Missingness is exogenous in these data;
no guarantee is claimed under outcome-dependent missingness. Restart the same
prior for every overlapping window; do not multiply recycled evidence twice.

Three arms at cadence8: strict current-version labels, latest64 labels across
versions, and frozen prior-only. Original experts publish every32 frames.
No current label is admitted. Under strict mode, reset at each publication.
Keep all 2688 runs, 21 cases, phases, schedules, all256 and terminal64 scores.
This baseline-only study cannot validate multi-expert correction on its own.

For each learning arm, compare generic64, Markov, matched cadence8 baseline-MAP,
and prior-only. Nonharm lower gain >= -.01 everywhere. Terminal gains on all
eight changing cases against generic64, Markov and prior-only require mean >=
.005 and lower > 0. Use paired mean +/- 3.5 SE over 32 trajectories, explicitly
exploratory not simultaneous. All gates required; no case-dependent selection.

Component verification: independent multiplicative enumeration, evidence,
finite extreme probabilities, no-data symmetry, ownership and invalid inputs.
Quality audit: poisoned unarrived labels, hidden q independence, prior-version
isolation for strict mode only. Record support, finite evidence and posterior
normalization. Reconstruct scores independently or replay byte-exact. No failed
trajectory may be silently discarded.

Costs: H=25 fixed hypotheses, N<=64 labels. Fit O(H*N), prediction O(H), memory
O(H+N). These are slow-path reference costs, not measured loaded serving latency.
Failure under both window modes rejects this discrete uncertainty treatment;
it does not prove that every Bayesian correction is impossible.
