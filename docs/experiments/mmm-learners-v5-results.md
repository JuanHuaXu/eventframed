# Learner refinement v5 results

Date: 2026-09-12. First executed research package for roadmap items 1, 2 and 4.
Protocol: [frozen design](mmm-learners-v5-protocol.md). Machine-readable evidence:
[summary](mmm-learners-v5-summary.json), [full journal](mmm-learners-v5.json.gz).

## Verdict

All three new mixtures FAILED the complete frozen acceptance rule. All passed
stationary protection and the no-large-mean-harm check on this finite matrix.
The forest is promising on early simple-rule shifts, not a generally validated
rescue. The adaptive-window replacement showed no useful consistent improvement.
No thresholds were retuned between design and confirmation.

This isolates learners using all nine input coordinates. It does not test MMM's
six-coordinate observation allocation, Anti-Pigeon sharing, text retrieval or
production serving. Scores must not be compared directly with v3/v4 as if their
observation budgets were the same.

## Confirmation Brier loss

Lower is better. Each row averages 24 streams (eight per fitting seed). Stable
and null rows use the full stream; other rows use the declared post-change window.

| Scenario | Fixed mixture | Adaptive mixture | Forest mixture | Combined mixture |
| --- | ---: | ---: | ---: | ---: |
| Stable, 5% noise | 0.063357 | 0.063356 | 0.063357 | 0.063357 |
| Stable, 20% noise | 0.177546 | 0.177547 | 0.177559 | 0.177559 |
| Shift at 128 | 0.249634 | 0.249248 | 0.186261 | 0.186202 |
| Shift at 256 | 0.254840 | 0.254990 | 0.245657 | 0.245503 |
| Shift at 384 | 0.263389 | 0.263864 | 0.262725 | 0.263206 |
| Gradual | 0.209857 | 0.210108 | 0.199351 | 0.199351 |
| Recurring | 0.200466 | 0.200710 | 0.192239 | 0.192529 |
| Delayed/missing labels | 0.266688 | 0.266681 | 0.265103 | 0.265465 |
| Interaction-only shift | 0.254136 | 0.254742 | 0.255941 | 0.255674 |
| Null | 0.252120 | 0.252169 | 0.252178 | 0.252178 |

Primary forest gains (fixed minus forest Brier):

- Shift128: 0.063373, approximate interval [0.037249, 0.089496].
- Shift256: 0.009183, approximate interval [-0.003018, 0.021384].

These use the frozen z=3.6 rule across 120 contrasts, conditional on the three
fitting models, not unconditional coverage over arbitrary training samples.
Both primary forest gains had positive means in each fitting-seed group, but
the later-shift lower bound crossed zero. Combined behaves similarly. Adaptive
missed the minimum gain and reversed sign on shift256 in all three fit groups.

Stable 5%-noise accuracy was about 94.80% for the forest mixture. Early-shift
accuracy improved from 53.07% (fixed mixture) to 74.15% (forest). This does not
mean 94.7% accuracy survives arbitrary shifts. The early-shift forest reached
the declared 80%-correct rolling-64 window in 23/24 streams versus 0/24 for the
fixed mixture; later shifts reached it in 9/24 (shift256) and 1/24 (shift384).
These are window-hit counts, not sustained-accuracy guarantees.

## Mechanism and limitations

Standalone short/adaptive models on stable05 had Brier 0.241709/0.239437 versus
the frozen incumbent's 0.062975. Preserving/mixing the incumbent is important in
these controls; throwing it away remains unsupported.

No adaptive cuts occurred in either stable confirmation scenario (24 streams
each). This finite observation is not a certified false-cut rate. Gradual also
had no cuts; recurring averaged 2.42. A cut alone did not fix sparse fitting.
Forest size reached at most 23 nodes in confirmation, below the 155-node cap.

All arms received the same audit evidence, but forests update each audit while
count models refit every 16 after initial support. Representation and update
cadence are therefore not separately identified. No forest-only arm was included.
The greedy tree has no forgetting/replacement and can struggle with interactions;
the XOR control and late shifts show why it is not ready for general adoption.

Missingness was independent (25%), with a fixed 16-step delay, not arbitrary
outcome-dependent missingness or out-of-order arrival. Missing labels were used
only for external scoring, never learning. Delayed/missing confirmation streams
averaged 373.08 delivered labels, 93.25 admitted audits and 11.79 pending packets.

## Cost boundary

Microbenchmarks are recorded separately in
[observable-result benchmark](mmm-learners-v5-benchmark-observable.txt).
The earlier benchmark file discarded predictions and is not authoritative for
prediction timing. Neither benchmark measures end-to-end latency, allocation of
observations, database I/O, background publication or concurrent serving.

Count-model refits were timed separately inside the experiment: approximately
465 microseconds per fit pooled across confirmation calls. This is a descriptive
mean over several model/window sizes, not a p99 bound. Forest updates and window
monitor costs are not included in that refit number. This research harness runs
synchronously; any future serving integration still requires background jobs.

## Next research step

Preserve this run. Freeze a cadence-matched comparison before new results, then
test the best supported candidate with partial-view MMM. A positive isolated
learner result is insufficient to modify the whitepaper's validated MMM claim or
the production daemon. No real-world generalization or deployment is established.
