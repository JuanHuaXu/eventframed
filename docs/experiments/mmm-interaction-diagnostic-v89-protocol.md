# Interaction diagnostic v89

Post-hoc replay of all2,560 v88 records, not new confirmation. No changes to
models, audits, observations, feedback, publications, selectors or randomness.
Use read-only pre-outcome model evaluations and verify exact parent-tape parity.

For eight64-step windows record emitted mixture Brier; hypothetical full-input
mixture Brier with the same weights/models; actual/full-input Brier for base,
short count, retained subset, age subset and active long model; availability;
outer/inner weights; guide identity; and coverage of the generator's required
variables. Missing model scores are not included in conditional averages.

Full inputs and the true required-variable mask are simulator-only diagnostic
information, never additional reads or training evidence. Required-mask coverage
is syntactic, not minimal sufficient information in the dependent-input cases.
Null has no required mask and its coverage denominator is zero. Pre-change and
stationary required mask is bits6/7/8; post-change masks follow declared rules.

Hypotheses: poor learned interaction model; missed joint observations; late
selector promotion. Actual/full-input comparisons are not causal deployment
gains, because changing observations could change subsequent online behavior.
Do not pick a rescue merely from the name of the failed parity case.

Require ordered prepare/score transitions, reconstruction of original mixture
and inner forecasts, bounded metrics, exact original records, source manifests,
focused race checks and full deterministic diagnostic replay. No runtime or
production change, no claim of performance from instrumented execution.
