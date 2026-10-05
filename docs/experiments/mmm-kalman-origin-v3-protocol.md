# Kalman origin-time update v3: frozen isolated protocol

Date: 2026-10-02. This Goal 2/4 study tests one explanation for the
delayed-recovery failure in v1/v2: a label generated at clock `o` is
currently assimilated into covariance at delivery clock `o+16`.
For the fixed-delay, in-order synthetic stream, the proposed filter
advances its working state only to the label's origin, updates there,
then predicts forward to each current clock. It remains a linear-Gaussian
working approximation to Bernoulli residuals, not an exact posterior or
an EventFrame serving change. The original v1/v2 records are retained.

## Frozen design

- Reuse the nine-bit generator and unchanged `runEarlyV2` control runner.
  Fit four independent frozen baselines per scenario using fit indices
  400..403 in design and 500..503 in confirmation; eight streams per fit.
  Stream seed bases are 2026102201 and 2026102202, disjoint from v1/v2.
- Scenarios: `stable05`, `shift128`, `gradual`, `delayed_missing`,
  `interaction`. Each split has 32 trajectories per case, 512 clocks each.
  Control arms are frozen baseline, last-64, delivery-time Kalman
  `q=.0005`, delivery-time `q=.002`, delivery-time `q=.01`.
- Add two origin-time filters with `q=.0005` and `q=.01`, identical
  features, initial state, observation variance and forecast clipping.
  Process noise is applied per origin-clock elapsed. Every arm sees the
  same contexts, potential outcomes, audit decisions, missingness and
  delivery schedule. Forecast at clock `t` precedes its outcome and all
  feedback delivered at `t`. Only arrived audited labels update filters.
- The fixed-delay/in-order assumption is checked on each stream. A label
  with origin later than the current clock, an out-of-order audited
  origin, or a nonfinite state fails the study. The first 16 post-change
  clocks in `delayed_missing` cannot identify the changed rule; report
  them separately from clocks 272..335, the first 64 clocks when a
  changed label can have arrived.

## Decision and measurements

Reconstruct per-trajectory Brier and report paired mean differences with
mean +/-3.5 SE across the 32 trajectories, separately for both frozen
splits. Origin-time handling passes this component only if at least one
predeclared `q` meets **all** of the following in each split:

1. Delayed post-availability Brier over clocks 272..335 improves over
   its same-`q` delivery-time control by at least .01, with a positive
   paired lower endpoint. It is no worse than last-64 by .005.
2. Stable full-stream harm relative to the frozen baseline has paired
   upper endpoint below .01. Interaction last128 harm relative to
   last-64 has paired upper endpoint below .01. Late gradual Brier is
   no worse than last-64 by .005.
3. In zero-delay `shift128`, origin and delivery forecasts agree within
   1e-10 at every clock. No future-data or state validity violation;
   isolated forecast/update p99 below 50 us and per-filter retained
   state at most 4 KiB.

Report delayed blind-16 and first64-after-change separately even if the
component passes. A component pass does not complete Goal 2 or Goal 4:
the fixed feature map, generator, selection scheme and front-end latency
remain unvalidated. Preserve design and untouched confirmation including
negative results. Do not tune `q`, change thresholds or select scenarios
after reading either split.
