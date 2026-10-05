# Partial-view learner v7 results

2026-09-12. [Protocol](mmm-partial-v7-protocol.md),
[summary](mmm-partial-v7-summary.json), [journal](mmm-partial-v7.json.gz).

480 streams, four arms, fresh fitting and evaluation seeds. Live inspection
is capped at six coordinates; independent full-field audits arrive only after
forecasts at equal acquisition volume. No production, real-text or AP-sharing
evaluation is claimed.

## Frozen verdict

All candidates FAILED. Rolling MMM passed primary-shift gains, stationary
protections and seed-sign consistency, but failed recurring-regime mean harm.
The earlier full-view v6 pass is not a successful complete MMM upgrade.

Confirmation Brier, 24 streams per case (stable/null full; other rows post):

| Scenario | Fixed MMM | Rolling MMM | Frozen guide | Breadth |
| --- | ---: | ---: | ---: | ---: |
| Stable05 | 0.049531 | 0.049529 | 0.049529 | 0.251619 |
| Stable20 | 0.163526 | 0.163528 | 0.163361 | 0.251766 |
| Shift128 | 0.203022 | 0.176881 | 0.255629 | 0.251646 |
| Shift256 | 0.238865 | 0.221025 | 0.258015 | 0.251497 |
| Shift384 | 0.263994 | 0.260089 | 0.263941 | 0.251183 |
| Gradual | 0.199230 | 0.192182 | 0.213830 | 0.251970 |
| Recurring | 0.189543 | 0.204290 | 0.196703 | 0.251513 |
| Delayed/missing | 0.268074 | 0.263927 | 0.271642 | 0.252338 |
| Interaction | 0.243027 | 0.251479 | 0.258464 | 0.251666 |
| Null | 0.251867 | 0.251671 | 0.251699 | 0.251285 |

Rolling primary Brier gains: 0.026141 [0.016252,0.036031] and
0.017840 [0.003503,0.032176], using the frozen approximate paired z=3.6 rule.
All fitting-group primary means improve. Intervals are conditional on the
three fitting models, not real-data generalization guarantees.

Stable05 accuracy is 94.82%. Shift128 improves from 67.24% to 75.72%; shift256
from 58.82% to 65.90%. Recurring declines from 69.06% to 66.49%. Recurring post
Brier harm is 0.014747, and full harm 0.011062, exceeding the frozen .01 limit.
Its approximate interval crosses zero: this fails the mean-based guard, not a
claim of statistically proven harm in every population.

## Interpretation

The frozen incumbent guide loses most shift adaptation. Breadth cannot reveal
the depth-dependent XOR/local-bit2 rules here; that does not make it generally
useless. Observation behavior materially affects this fixture's outcomes.

Replacing the short-count slot changes the learner and the guide. Losing that
count-model capability during regime returns is a plausible cause of recurring
and interaction weakness, not yet an isolated causal conclusion. Next freeze an
additive short-count/tree challenger mixture with predictable weighting, keeping
the incumbent AND the short-count option. Retain all adverse scenarios and the
same live budget; do not widen the harm threshold or overwrite v7.

## Verification

Unit/race checks passed for uniform marginalization versus enumeration,
hidden-value invariance, budget6, invalid-reader rejection and missing/delayed
feedback exclusion. Complete outcomes, views, forecasts and verdicts replayed
exactly except wall-clock timings; source/protocol hashes match. Vet passed.

Uniform tree marginalization is justified only by the declared independent
fair-bit generator. Real EventFrame fields need another validated observation
model. Hidden current-row values never enter a live forecast; full fields enter
training only through the separately declared available audits. No production
latency or scalability claim is established by this run.
