# Predictor-version-local calibration diagnostic

Frozen before scoring this candidate. Consumed v120 data only; no fresh
confirmation, production changes, or inherited calibration guarantee.

Root-cause status: timing mismatch confirmed; attribution to predictor drift
alone needs investigation. Alternatives include environmental shift and
finite-sample overfit. This experiment restricts reuse, not the base learner.

All eight selected original experts publish on clocks divisible by 32. Define
version(t)=floor(t/32). Fit the existing unit-prior residual-logit model using
only earlier arrived nonmissing labels whose ORIGINAL issued forecast belongs
to version(t). Reset to the zero-correction prior at publication. Never
hindsight-refit training forecasts. Current-frame outcomes are unavailable.

Four fixed candidates: baseline-only at cadence 8, baseline-only at cadence 16,
full eight-expert at cadence 8, full eight-expert at cadence 16. All fits restart
from the original prior. Within-version support is at most 24 labels at a fit;
the 64/32 old caps would be redundant and are not compared. Cadences measure
the support/update-cost tradeoff, not a post-result optimization. This is an
offline slow-path reference, not permission to fit in the serving path.

Use all 2688 original runs, all 21 cases, both phases/schedules, full256 and
terminal64 scores, expected Brier, expected accuracy and expected log loss.
Retain generic64 and Markov controls for each candidate, and baseline-only,
both variational caps and both segment caps for full candidates. Gain gates
apply to generic, Markov and matched baseline calibration on the eight changed
cases' terminal64. Nonharm lower gain >= -.01; gain mean >= .005 and lower > 0.
Intervals are mean +/- 3.5 SE over 32 trajectories: exploratory, not simultaneous.
Every gate must pass for qualification; identity-like fallback alone cannot pass
the gain requirement. No candidate selection on teacher q or case identity.

Record every fit, support and forecast score. Verify prefix independence under
unarrived-label/teacher poisoning, and independence from prior-version labels
while holding original expert forecasts fixed. The latter tests the correction,
not independence of the underlying learner from its legitimate training history.
Numerical failure aborts rather than dropping a trajectory. Compare support
and fit counts, but do not infer loaded performance from diagnostic wall time.

Falsifier: if same-version correction still harms stable cases or cannot recover
both switches, strict reset does not rescue the workload. Improved cells alone
do not validate the seven research goals.
