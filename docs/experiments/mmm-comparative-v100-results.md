# v100: comparative forecast falsification

## Verdict: FAIL, 104/106 quality gates

All48 whole-stream and48 late-half non-harm gates pass, as do all6 stationary
interaction-gain gates. Both majority-to-parity recovery gates pass. Both
parity-to-majority positive-recovery gates still fail. All four finite null
checks pass. More rejections did not rescue the remaining direction.

The [frozen protocol](mmm-comparative-v100-protocol.md) retains ten controls and
adds the comparative gate on768 fresh learned streams. Each target forecast is
tested against a fixed mixture of neutrality and the three other raw forecasts.
The64-test allocation, threshold6400, scheduled versions and total evidence
volume stay unchanged. Crossing affects the NEXT prediction, not the revealing
outcome's already-issued prediction. Raw models and bank weights keep learning
identically under both gates. No production path was changed.

See [raw records and null counts](mmm-comparative-v100.json),
[all gates and control metrics](mmm-comparative-v100-summary.json), and
[independent evaluator](../../research/comparative-v100-summary.mjs).

## Matched confirmation results

| Full-view scenario and segment | Generic64 Brier | Raw bank | Neutral gate | Comparative gate |
| --- | ---: | ---: | ---: | ---: |
| Parity4, all256 | 0.096817 | 0.076964 | 0.076968 | 0.075500 |
| Parity4, late128 | 0.072028 | 0.053689 | 0.053697 | 0.053688 |
| Majority to parity, late128 | 0.193750 | 0.188111 | 0.180430 | 0.173365 |
| Parity to majority, late128 | 0.178977 | 0.183201 | 0.176421 | 0.176502 |

Parity4 late expected accuracy remains94.78%. This is a noisy finite Boolean
task, not measured real-world agent accuracy. Comparative majority-to-parity
gain over generic64 is0.020385, interval[0.012227,0.028544]. Its gain over
neutral-only is0.007066 in the matched mean; this comparison is descriptive,
not a new predeclared success gate.

Parity-to-majority gain over generic64 is only0.002474,
interval[-0.001500,0.006448], below the0.005 mean requirement and with a lower
bound below zero. Design fails too:0.002765, interval[-0.002302,0.007833].
Comparative Brier is0.000081 worse than neutral-only on confirmation in this
direction. Do not infer improvement by comparing these values with different
v99 samples. Intervals are the declared paired32-trajectory z=3.5 approximation,
not exact coverage or anytime confidence sequences.

## Rejection is not promotion

Mean full-view masked frames per256-step confirmation stream:

| Scenario | Neutral gate | Comparative gate |
| --- | ---: | ---: |
| Parity4 | 0.375 | 13.000 |
| Majority to parity | 13.250 | 29.500 |
| Parity to majority | 19.0625 | 32.46875 |

The extra masks are not known false rejections: learned forecasts need not
satisfy the conditional calibration null. Comparative neutral-fallback means
are0,0.4375 and0.21875 frames respectively. Rejection timing reconstructs every
mask/fallback count, with a strict preceding-outcome condition.

In parity-to-majority confirmation, generic32's late Brier is0.155822, better
than generic64's0.178977, yet its raw weight at step160 averages0.006381.
Generic64 still carries0.707566, with the two Boolean experts holding the rest.
Comparative testing rejects generic64 in only5/32 streams during that version,
versus15/32 for Boolean64 and8/32 for Boolean32. Generic32 is never rejected
in that version. Survivor renormalization retains old relative weights; these
consumed diagnostics motivate testing evidence-attributed routing, but do not
prove underweighting is the only cause. Stale forecasts before refitting and
short-window variance after the long model catches up remain limitations.

## Known-null simulations

| Mode | Monitor | Families with any rejection / families | Wilson95% upper |
| --- | --- | ---: | ---: |
| Independent target tests | Neutral | 7/4096 | 0.3524% |
| Independent target tests | Comparative | 6/4096 | 0.3192% |
| Shared within version | Neutral | 0/4096 | 0.0937% |
| Shared within version | Comparative | 0/4096 | 0.0937% |

Each mode uses paired outcomes for both monitors. Independent mode includes
predictable history-dependent alternatives; shared mode gives all four experts
and both views the same calibrated law. All four upper bounds are below1%.
These are finite simulation screens, not proof of calibration for the fitted
models, guaranteed future accuracy, causal validity or poisoning detection.

## Verification and performance

- Learned race smoke passes:3.515s package time.
- Literal/reference, adaptive-null, lifecycle and raw-bank race checks pass:
  1.286s package time. Vet passes.
- Full generation passes:141.313s; complete learned/null replay passes:142.728s.
- All19 source hashes, independent summary reproduction, publication origins,
  raw-bank bounds and both gates' rejection-count reconstruction pass.
- Artifact SHA256:fea755669d238f8b7978e6ad85785123299f718d61747f3ff201962168e35a28.

The [whole eleven-arm fixture](mmm-comparative-v100-benchmarks.txt) takes
177.30-198.06ms for256 scored steps, allocating21.55-21.58MB across2059-2191
allocations. Three single-iteration repetitions ran after the experiment and
replay finished, on Apple M4,Go1.27.1 darwin/arm64,GOMAXPROCS10. This includes
all constructed models, eight refits, both views, scoring and hashing. It is
not per-request serving latency and is not a matched production overhead test.

The earlier [paired component benchmark](mmm-comparative-v100-component-results.md)
measures1.105-1.123us per comparative predict/update pair versus0.221-0.224us
for neutral-only, with zero allocations. The richer gate costs about five times
the monitor arithmetic but does not fix the remaining recovery failure.

Next: [evidence-attributed routing](../../research/evidence-routing-proposal.md),
with separate fresh validation and unchanged gates. The raw bank's cumulative
loss guarantee does not cover either gated or proposed routed forecast. No
whole research direction is completed or exhausted by this result.
