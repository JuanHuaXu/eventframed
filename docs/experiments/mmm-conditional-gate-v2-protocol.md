# Conditional-gate forecast transfer, frozen protocol

Date: 2026-10-01. Freeze before running new seeds. V1's 4,000 streams are
consumed design evidence; do not reuse their outcomes for this forecast test.
Use seed base `2026100191`, 1,000 independent trials per regime and schedule.
Keep the v1 two-cell world, 256 independent reference labels per cell, 512
live frames, scalar/conditional threshold, `delta=0.02`, `epsilon=0.10`, and
complete-immediate versus 25%-audit/20%-missing/delay-0..31 schedules.

Both arms retain the same `(X,Y)` evidence when an externally nominated label
arrives. Only split *authorization* differs: one gate projects the stream to
scalar correctness, the other checks each declared context. The reference
forecast for context `x` is the Beta(1,1) mean
`(1+reference_success[x])/(2+256)`. Both arms issue this law until their own
gate flags. Thereafter each uses the Beta(1,1) mean from its own already-
arrived live audit labels in that context:
`(1+live_success[x])/(2+live_count[x])`. This is a declared working
conditional estimator, not a claim of calibrated Bayesian belief under
selective or dependent observations. Evidence before a flag may be reused
after authorization; no outcome from the current or future frame can update
its own issued forecast. Neither arm receives the synthetic target law.

Primary outcome: paired, trajectory-level mean realized Brier gain
`scalar - conditional` over all 512 live frames. Also report expected Brier
under the hidden generating law *only in the evaluator*, accuracy, audit
counts, split clocks, and full-trial runtime. The two arms have identical
reference, live inputs/outcomes, audits, delays and missingness within each
trial. They differ only in gate decision. No model fitting beyond the declared
four counters; no additional acquisition by the conditional arm.

Predeclared success screens, separately for each schedule:

- Stable regime: each gate flags at most 20/1,000; paired realized Brier harm
  upper endpoint (mean plus 3.5 trajectory SE) is <=0.01.
- Swapped regime: conditional gate flags at least 900/1,000, scalar gate at
  most 20/1,000; paired realized Brier gain mean is >=0.05 and mean minus
  3.5 trajectory SE is >0.
- Exact same-tick issued probabilities must be calculated before event
  outcome generation or delivery. Labels with arrival clock > decision clock,
  missing labels, and duplicate origins must fail closed. Every observed
  update must be accounted for by an arrived audit in both arms.
- Preserve all 4,000 trial records and failures. Recompute summaries from raw
  rows in a separate verifier; deterministic replay should match records
  apart from elapsed runtime. No threshold or seed changes after inspection.

These finite trajectory intervals are descriptive experiment gates, not
simultaneous target-law coverage or an anytime causal certificate. Passing
could establish a *toy downstream benefit* of conditional authorization,
not Goal 3 completion or EventFrame production readiness. The reference
sample costs 512 labels and must not be hidden from resource accounting.
