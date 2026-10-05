# Kalman residual challenger v1: frozen protocol

This is an isolated Goal 2/4 component screen, not an EventFrame integration or
an ordinary Bayesian posterior for Bernoulli labels. The candidate is a
linear-Gaussian working filter for the residual of a frozen, pre-change
observation-model forecast. It may fail when the relevant relation is nonlinear.

## Data and separation

- Use `observationlearners`' existing nine-bit generator and `observation.Fit`
  base. Cases are `stable05`, `shift128`, `gradual`, `delayed_missing`, and
  `interaction`, each for 512 forecast-before-outcome ticks.
- Design and confirmation use independent stream and fitting seeds, four fits
  times eight streams per case and split. Inputs, audit flags (25%), missing
  flags, outcomes, and delivery schedule are paired across every arm.
- Only arrived audited labels update candidate and window forecasts. The
  existing loss-window detector may use all arrived labels, as in its current
  design. No issued prediction is revised after its outcome arrives.
- A delayed label is applied at delivery time to its origin-time feature row.
  This is an explicit, approximate working filter, not an exact out-of-sequence
  Kalman smoother. Results in delayed streams must be interpreted accordingly.

## Frozen arms

1. Frozen pre-change base model.
2. Last-64-audit model, refit after audit counts 32, 48, ... .
3. Existing capped loss-window model, refit on retained audits on the same
   cadence. Its detector receives arrived losses of issued base predictions.
4. Kalman residual: eleven coordinates (intercept, centered base probability,
   nine signed context bits scaled by 1/3). Initial mean zero and diagonal
   covariance 1; random-walk process noise 0.0005 per coordinate per tick;
   observation variance 0.25; output clipped to [0.01, 0.99]. Its working
   observation is `label - base_probability`. No data-dependent retuning.

The three fitted/filtered challengers see the same audited labels. The Kalman
arm predicts every tick, advances covariance once per tick, and updates only
when a label is available. Its feature map is fixed before the split.

## Measures and decision

Report full and last-128-tick Brier/accuracy, paired per-stream Brier gains,
first 64 ticks after each change, audit and delivery counts, finite-state and
positive-covariance checks, estimated resident bytes, and isolated update
latency. Paired intervals are mean +/- 3.5 standard errors over the 32
trajectories and are descriptive screens, not simultaneous coverage claims.

The candidate passes this screen only if both `shift128` and `gradual` show
last-128 Brier gain of at least 0.005 over the better of the two window arms
with a positive paired lower endpoint; the `stable05` full-stream harm against
the frozen base and the better window has an upper endpoint below 0.01; and
`delayed_missing` last-128 harm against the better window has an upper
endpoint below 0.01. The `interaction` case is a declared nonlinear negative
control and must be reported, not hidden. The update p99 must be under 50 us
in an isolated one-thread component benchmark and state under 4 KiB.

Passing is only evidence for these synthetic generators. It would not prove
Anti-Pigeon validity, improved agent answers, actual serving p99, or Goal 2/4
completion. Keep the design result and confirmation result even on failure.
