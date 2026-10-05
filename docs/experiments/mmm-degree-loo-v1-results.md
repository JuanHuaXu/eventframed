# Degree LOO v1: full results

## Verdict

All four candidates FAIL the unchanged broad quality gates. This is consumed
synthetic evidence, not fresh confirmation or live agent data. No promotion,
whitepaper adoption, or completion of any of the seven research goals.

| Objective / labels | Non-harm | Recovery gain | Result |
| --- | ---: | ---: | --- |
| Smooth squared /64 | 254/840 | 8/128 | FAIL |
| Smooth squared /32 | 250/840 | 11/128 | FAIL |
| Clipped Brier /64 | 232/840 | 8/128 | FAIL |
| Clipped Brier /32 | 243/840 | 11/128 | FAIL |

The protocol, original five comparator families, sample counts and thresholds
are unchanged. No partial-run outcome informed the fit settings. The scorer
is the unchanged `research/spectral-regression-summary.mjs`.

## Compare The Previous Objective

Matched comparisons against the old16-step fixed-noise marginal-likelihood
fits preserve both scoring spans, all phases/cases/schedules and32 trajectories
per cell. Positive means lower Brier for LOO. These supplementary contrasts
do not replace the broad gates.

| Objective / labels | Non-harm vs old | .005 positive-lower gains | Mean gain |
| --- | ---: | ---: | ---: |
| Smooth /64 | 79/168 | 67/168 | +0.00262935 |
| Smooth /32 | 68/168 | 48/168 | -0.00652357 |
| Clipped /64 | 80/168 | 68/168 | +0.00124159 |
| Clipped /32 | 64/168 | 47/168 | -0.00722165 |

The64-label models have small positive descriptive averages but fail protection
in many cells; the32-label averages worsen. Overlapping spans are not independent
replicates. This is not evidence that LOO never helps, or that a proper score
is a bad objective. It shows this bounded-window hyperparameter-fitting method
does not deliver the required broad improvement. Lower fitted LOO error does
not supply independent information or establish future calibration.

Selected phase1 delayed-feedback terminal Brier values illustrate the trade-off:

| Case | Smooth64 | Clipped64 | Original Markov |
| --- | ---: | ---: | ---: |
| Additive stationary (0) | .23289 | .23689 | .22176 |
| Additive gradual (2) | .25332 | .25613 | .23907 |
| Hierarchical gradual (5) | .26812 | .27260 | .24446 |
| Local-table gradual (8) | .27952 | .28627 | .25039 |
| Parity4 (12) | .19272 | .19352 | .04907 |
| Majority to parity (19) | .23036 | .23304 | .06026 |
| Parity to majority (20) | .14456 | .14239 | .10227 |

The previous fixed-noise64 parity4 Brier was .22825, so LOO improves it but is
still far behind the retained Boolean/Markov machinery. Accuracy there is
about70%, not the existing specialist's95%. These are illustrative cells;
all cells and paired intervals remain in the machine-readable artifacts.

## Numerical Checks

2688 records and129024 total fits complete, including86016 LOO fits.
85275/86016 meet projected-gradient tolerance;741 hit the iteration cap.
No numerical errors or line-search caps. All finite fitting objectives are
nonincreasing; this is a solver check, not a quality guarantee.

| Objective / labels | Converged | Capped | Mean evaluations |
| --- | ---: | ---: | ---: |
| Smooth /64 | 21247 | 257 | 84.36 |
| Smooth /32 | 21404 | 100 | 77.00 |
| Clipped /64 | 21218 | 286 | 83.03 |
| Clipped /32 | 21406 | 98 | 76.83 |

The full-record auditor checks all86016 optimizer records. Independent
Gauss-Jordan reconstruction checks1008 stratified fits,2016 objective values,
32256 forecasts and4032 finite-difference derivatives. Maximum errors are
3.24e-14 for predictions,4.44e-16 for objectives and1.77e-11 for projected
gradient norms.1376256 linear predictions agree exactly with the older tape.
The pilot's race, clipping-branch and model-level future/missing-label tests
remain applicable; the implementation and adapter were not changed afterward.

## Cost And Next Lead

Full collection took1305.93s test-body time; full replay1306.63s and is
byte-identical across all2688 records. The isolated64-label fixtures
cost38-42ms per fit, versus roughly4ms for the old capped method. These are
slow-path fitting costs, not request latency. No new serving-speed measurement
or production performance claim follows.

Do not tune another LOO penalty or mixer prior on these results. First inspect
whether degree-wide variance sharing is the representational bottleneck:
it assigns the same variance to all interactions of a given order, whereas
the retained Boolean specialist can isolate an individual parity. A sparse
multi-interaction learner is a distinct possible lead, not a proven rescue;
check existing implementations and primary research before proposing its
mathematics. [Initial source inspection](../../research/sparse-interaction-source-note.md)
records Tipping2001's classification mechanism and the author's corrected
working-response equation. It is not yet an implementation proposal or proof
of benefit. Keep the strong existing specialists and all negative controls.

## Artifacts

- Protocol and component evidence: `mmm-degree-loo-v1-protocol.md` and
  `mmm-degree-loo-v1-pilot-results.md`.
- New full forecasts: `mmm-degree-loo-v1-forecasts.jsonl`.
- Quality gates: `mmm-degree-loo-v1-summary.json`.
- Numerical audit: `mmm-degree-loo-v1-audit.json`.
- Old-objective contrasts: `mmm-degree-loo-v1-contrasts.json`.
- Full run log: `mmm-degree-loo-v1-run.txt`.
- Audit source: `research/degree-loo-v1-full-audit.mjs`, a detached extension
  of the preserved pilot auditor; contrast source:
  `research/degree-loo-v1-contrasts.mjs`.

Full replay artifact: `mmm-degree-loo-v1-replay.jsonl`, with its separate
`mmm-degree-loo-v1-replay-run.txt` log. Exact byte comparison passes.

SHA256 forecasts/replay:
`3f7231f806b17f89ecd33241c96db8ff5f47e410174d85fbbba5a6780e374819`.
Summary:
`e9720c131a916c793a254f37d7c6afb7ad549b5405ee62248675ce68689681f3`.
Audit:
`25eefc4835e68f16e63144e1d828dd2f443d59a09673cc53b7ea86cbd4263dd4`.
Contrasts:
`074c4f1b73c654e96e91537920c0e7884841cd3b703abcb4f2fab4b3e6cd22eb`.
Implementation and adapter hashes still match the pilot record.
