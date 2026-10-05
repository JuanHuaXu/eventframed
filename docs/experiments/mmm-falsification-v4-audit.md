# Paired MMM v4 audit

- The approved test changes evidence selection only; the v3 preserved forecast
  state and sharing gate were imported unchanged. Only research package/command
  and evidence files were added. No production configuration or service changed.
- Unit/race checks passed before evaluation: seed separation, uncertainty versus
  disagreement choices, independent exploration, identical-observer negative
  control, next-step fitting, and acquisition-cost accounting.
- All choices were fixed before outcomes became available. Same candidate means
  same potential outcome across arms. Unchosen labels do not enter fitting.
  Live outcomes, not selected probe outcomes, update forecast-mixture weights.
- Source and protocol hashes match. All scores and 60 contrasts recomputed.
  Every stream, candidate choice, observed mask and forecast replayed identically,
  excluding only measured fitting/append wall time. Full replay ran without race;
  the separate unit/one-stream race check passed. `go vet` passed.
- A missing closing brace in the newly added replay-test source was repaired;
  this did not affect the already compiled experiment, frozen sources or results.
- The paired selector chose different evidence on over half of relevant audit
  opportunities. The negative outcome is not explained by an inert selector.
- Selection has known positive support but is not corrected into an ordinary
  Bayesian posterior. Independent live scores are the evaluation boundary.
- Limitations: synthetic supplied views, bounded candidate pool, one fitting seed,
  small number of streams, no real-data or field-level discovery test, no unique
  Anti-Pigeon ablation, no production performance measurement. All settings and
  negative results retained; no outcome-driven tuning or seed replacement.
