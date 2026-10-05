# Paired MMM falsification v4 results

## Verdict

**The added-value criterion did not pass.** Paired disagreement-guided evidence
selection stayed within the stationary harm ceiling but did not reliably improve
on random or single-MMM uncertainty selection on either single-shift scenario.
The preserved-incumbent recovery from v3 remains intact. This is a negative result
for this evidence-selection heuristic, not a rejection of all falsification-based
observation or of MMM's previously measured observation-allocation benefit.

The tested interpretation is bounded candidate-event selection for outcome queries.
It is not autonomous design of field-level observations, causal interventions,
or explicit alternative mechanisms. The two observers use the highest-weight
available predictors from the existing family; they do not invent new hypotheses.

## Evidence

Frozen protocol: `mmm-falsification-v4-protocol.md`.
Raw predictions, probe inputs, observer masks, choices and subsequently revealed
labels: `mmm-falsification-v4.json.gz`.
Per-stream scores, all 60 contrasts, recovery windows and hashes:
`mmm-falsification-v4-summary.json`.

Two splits x five scenarios x 32 independent streams x 512 live steps = 163,840
live evaluations, or 491,520 paired policy forecasts. The latter are not
independent observations. Three arms had the same probe pool and per-step
acquisition allowance. Initial fit and all evaluation RNG domains were separate;
no settings were changed after seeing the design or confirmation results.

## Confirmation Brier

Lower is better. Stable/null use full512; single shifts use post256; recurring
uses the 384 steps from its first change. All learning warm-up is included.

| Scenario / window | Random | Single MMM | Paired MMM |
| --- | ---: | ---: | ---: |
| Stable / full | 0.050370 | 0.050258 | 0.050272 |
| Member shift / post | 0.237473 | 0.239564 | 0.237197 |
| Common shift / post | 0.241657 | 0.240821 | 0.242681 |
| Recurring / post | 0.191887 | 0.192116 | 0.194963 |
| Null / full | 0.251898 | 0.251793 | 0.251870 |

Paired gain over random was 0.000276 [-0.007101, 0.007652] on member shift
and -0.001024 [-0.008299, 0.006251] on common shift. Against single MMM it
was 0.002367 [-0.005123, 0.009857] and -0.001860 [-0.009193, 0.005473].
None met the required gain of 0.005 with a positive paired lower bound.

These are the frozen approximate simultaneous z=3.5 paired normal intervals over
32 stream means, conditional on the fitting model. Failure to meet the improvement
criterion does not prove the effects are exactly zero. Recurring paired gain
over random was -0.003076 [-0.009937, 0.003786], also unresolved. No confirmation
scenario/window exceeded the 0.01 mean-harm threshold against random.

## Learning speed

Recovery means reaching at least 80% correct in a fully post-change rolling
64-step live window. This is descriptive, not a stopping-time certificate.

| Scenario | Random recovered /32 | Single /32 | Paired /32 |
| --- | ---: | ---: | ---: |
| Member shift | 6 | 7 | 8 |
| Common shift | 6 | 8 | 3 |

Paired therefore missed recovery in 24/32 and 29/32 streams. Among its detected
subsets, mean window-completion delays were 238.25 and 244.67 steps from the
change. Random's detected-subset means were 216.67 and 243.67; different subsets
make those means unsuitable as a stand-alone speed comparison. There is no
consistent demonstrated learning-speed advantage.

Post-change paired accuracy was 59.33% / 56.98%, versus random 59.59% / 58.18%.
Late(last128) paired Brier was 0.212856 / 0.219510 versus random
0.212091 / 0.217836. The late window does not reveal a hidden recovery win.

## Was the observer active?

Yes. In post-change audit opportunities, paired selection chose a different
candidate from random on 56.89% of member-shift and 55.50% of common-shift
opportunities. It differed from single MMM on 55.52% / 56.81%. Mean absolute
disagreement between its selected observers was 0.3665 / 0.3986. This rules out
a simple no-op selector, but disagreement alone did not yield more useful
training evidence with the tested predictor family and budgets.

The selection distributions have at least 1/16 probability per candidate through
independent uniform exploration. That does not make the selected stream unbiased:
the conditional-count challengers are explicitly working estimates, with possible
projection bias from selective input coverage. Their influence was evaluated on
independent live cases, not scored retrospectively on selected training examples.

## Sharing and cost

Stable, common-shift and null confirmation had zero splits; member and recurring
had 32/32. Every arm uses the same full-stream reliability gate, so this is not
an Anti-Pigeon ablation. Disagreement never authorizes a split. Stable zero out
of32 has Wilson95 upper bound 10.72%, not empirical proof of a zero false-split
population rate. The v3 conditional sequential guarantee has its own assumptions.

All arms were charged 36 input-coordinate units and two outcome queries per
audit opportunity. With the .25 opportunity rate, confirmation used about
8.87--9.16 probe-coordinate units and .493--.509 outcome queries per live step.
Informative foreground means were approximately4.00--4.89 coordinates, with
eight separate diagnostic coordinates; null used six plus twelve diagnostics.
Both observers' predictions were instrumented even in controls. These counts
equalize evidence availability; they do not establish a CPU or latency benefit
over a naturally cheap random sampler. No new serving-latency claim is made.

## Scope of the negative result

This tested predictive disagreement over a small candidate pool, not a mechanism
that explicitly asks which upstream variable would falsify a proposed causal
chain. Two predictors from the same family may share blind spots, and large
disagreement may reflect uncertainty or mismatched views rather than a useful
distinguishing variable. Those are plausible explanations, not proven causes.
An explicit contrastive-hypothesis/observation-design experiment would be a
different test requiring its own protocol and fresh confirmation data. No such
upgrade is claimed here. Nothing was deployed, pushed, or added to the whitepaper.
