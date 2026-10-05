# Hazard-mixture v113 results

Status: FAIL. No candidate promotion or production change.

The frozen comparison completed 1,536 schedule-runs over 768 latent
trajectories, with nine paired policies. Candidate 8 marginalizes eleven
switching-rate hypotheses; it does not retrospectively select a winning rate.
See [protocol](mmm-hazard-v113-protocol.md) and
[machine-readable summary](mmm-hazard-v113-summary.json).

## Required gates

- 804/818 pass overall.
- 767/768 non-harm gates pass.
- 37/50 gain gates pass.
- All 14 failures concern the late segment of delayed/missing switch cases.

The failed non-harm screen is design majority-to-parity against fixed Markov:
mean Brier harm .006251, interval [-.001183, .013686]. The upper limit exceeds
the frozen .01 allowance. This is failure to establish non-harm, not proof of
positive harm. Gain failures and their complete intervals remain in the summary.
Intervals are paired mean +/- 3.5 standard errors over 32 trajectories,
approximate fixed-sample screens, not confidence sequences or guarantees across
the accumulated research history.

## Confirmation, delayed/missing, late segment

Lower Brier is better. These are descriptive cells, not substitute criteria.

| Scenario | Generic | Arrival log | Fixed Markov | Rate mixture |
| --- | ---: | ---: | ---: | ---: |
| Parity4 | .073409 | .049750 | .049458 | .050133 |
| Majority to parity | .253930 | .218414 | .213236 | .213365 |
| Parity to majority | .219070 | .220992 | .218729 | .214322 |

Rate-mixture expected accuracy is 95.00%, 69.38%, and 72.19%, respectively.
The 95% result is a synthetic stationary cell at the fixture's noise ceiling,
not a general chatbot accuracy claim. Against fixed Markov, reverse-switch
gain is .004407 [.000812, .008001], below the .005 mean-gain requirement.
Forward-switch gain is -.000129 [-.008534, .008276].

## Audit trail

- Generation passed in 229.79 seconds.
- All 44 frozen source hashes verified.
- Independent JavaScript reconstruction verifies Brier/log scores, accuracy,
  observation costs, as-of training origins, selector clocks, ring accounting,
  paired latent streams and effective-seed uniqueness.
- Recomputed summary is byte-identical.
- Artifact SHA256:
  `54b6220b6936227e4e1e45954a67522a5f7eb152adb32a3bc8ef5f5283550860`.

Full deterministic replay passed in 231.42 seconds. The pre-generation
compatibility/effective-seed tests passed under the race detector; vet passed.

The [whole-fixture benchmark](mmm-hazard-v113-benchmarks.txt) retains all three
repeats per schedule on Apple M4: immediate 137.02-159.72ms, delayed/missing
168.01-169.80ms, approximately 27.3MB allocated per operation. Each operation
contains 256 frames, training fits and all nine policies. This is neither
per-query latency nor a loaded serving/p99 measurement. The benchmark used a
previously consumed v112 design fixture, not a selected confirmation trajectory.

## Interpretation and next investigation

Uncertainty over a constant switching rate has not established the requested
recovery gain. Correct component inference does not imply useful policy quality.
The earlier isolated lifecycle benchmark also showed substantial extra delayed
filtering cost; the present quality result does not justify deployment.

Next diagnose the consumed switch trajectories using a declared decomposition:
current model error, weighting error at identical observations, and observation
selection error. Compare both switch directions and stationary controls. An
oracle diagnostic may locate a limit but must not become an implementable claim.
Do not tune rates, priors, thresholds or sample size on this confirmation set.
Any subsequent rescue needs a fresh frozen comparison. All seven research
directions remain open, including independent generators and real prospective
tasks.
