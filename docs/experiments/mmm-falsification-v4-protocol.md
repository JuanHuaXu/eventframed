# Paired MMM falsification v4 protocol

Frozen before v4 evaluation, 2026-09-12. Scope: active selection of evidence
events from a bounded synthetic probe pool. This is NOT autonomous field-level
investigation or an SCM intervention. Preserve v3's forecast mixer, incumbent,
short/long models and sharing gate; change only which probe is admitted for
learning. No settings change after evaluation results are viewed.

## Arms and timing

Three arms: random, single_mmm, paired_mmm. Each emits the same v3-style
preserved-MMM forecast on an independently drawn live frame, before outcomes.
At an independent Bernoulli(.25) audit opportunity, materialize four probe
EventFrames, all nine binary inputs each, without outcomes. All arms pay 36
input-coordinate units and request ONE live and ONE reference outcome for the
selected probe. This intentionally equalizes input acquisition and isolates
allocation of expensive outcome evidence, not the cost of naturally blind random
sampling. Actual live/reference diagnostic and foreground costs are separate.

Each arm selects its two highest previous-weight available non-uniform models
(original, short and appropriate local/pooled long; ties original first). Each
model independently runs the existing MMM scope/depth observer on each candidate
using its six-coordinate internal inspection cap. Missing second model emits .5.
They may attend different views; disagreement is not a causal finding.

- random: uniformly choose one of four probes.
- single_mmm: choose maximal binary predictive entropy of the first observer.
- paired_mmm: choose maximal Jensen-Shannon divergence of the two Bernoulli
  forecasts, H((p+q)/2) - (H(p)+H(q))/2.

Ties use one shared random number, uniformly among maximizing candidates.
All targeted arms use 25% independent uniform exploration (same coin/number).
Thus every probe has nomination probability >= 1/16. Record chosen probability,
candidate inputs, both forecasts, scores, choice and subsequently observed labels.
No outcome is generated or exposed to a selector until all choices are frozen.
Shared candidates have identical potential outcomes across arms; unselected
outcomes never enter fitting. Repeated observation never increases sample count.

After live forecasts and probe choices, reveal independent live/reference
outcomes. Update mixture weights ONLY with the earlier live forecasts and live
outcome, not selected probes. Add only the selected probe pair to that arm's
audit buffers. Fit short64 / local256 / pooled128+128 at 32 pairs, then every16.
Original fit never changes. Selected probe conditional-count models are WORKING
predictors: selection depends on inputs and history, and partial projections may
be biased by the changed input distribution. No ordinary-posterior or calibration
claim; independently sampled live scores test usefulness and harm directly.

Anti-Pigeon nomination and v3's bounded sequential reliability gate receive
unselected full-stream original-model correctness. Observer disagreement NEVER
authorizes splitting. No-AP ablation is outside this test; v3's unresolved AP
increment remains unresolved.

## Data and seeds

Five v3 scenarios, 32 streams x 512 steps x two splits: stable process XOR,
member shift to local bit at256, common shift at256, recurring every128, fair null.
Five-percent independent label noise except null. Fit4096 examples with seed
2026092001. Design/confirmation bases 2026092002/2026092003, using v3's disjoint
base*1000000 + scenario*100000 + stream*100 + role convention. Roles0/1 live and
reference,2 audit,3 probe inputs,4 exploration,5 tie/choice,6..13 independent
candidate-pair outcomes. Initial fitting, live observations and probes are separate.
No reuse of v1/v2/v3 seeds; scopes supplied, not inferred from natural language.

## Criteria and interpretation

Primary added-value claim: paired_mmm post-change Brier gain >= .005 over BOTH
random and single_mmm, positive approximate simultaneous lower bounds, on BOTH
member/common shift confirmation. Stationary harm upper bound <= .01 against
random, plus stable empirical split fraction <= .05. No overall pass otherwise.
Compare full, post and late(last128) Brier; log loss, accuracy and confident errors.
Report every scenario, including null and recurring. A gain over old frozen MMM
is not evidence of benefit over v3-style recovery and cannot rescue this criterion.

For learning speed report independent-live cumulative correct counts and the first
post-shift 64-step window with >=80% accuracy, or a miss. This window measure is
descriptive, not a sequential certification procedure, and applies to single shifts
only. Report misses rather than treating non-recovery as a zero delay.

Use z=3.5 paired normal intervals over32 independent streams, conditional on the
fitting model. Two controls x3windows x5scenarios x2splits=60 predeclared contrasts.
No tuning, selective seed replacement or outcome-based early stopping. Preserve
all live forecasts and selected-probe traces, hashes and deterministic replay.
These are empirical fixed-sample intervals, not confidence sequences.
