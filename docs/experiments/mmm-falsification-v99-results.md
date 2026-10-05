# v99: version-scoped forecast falsification

## Verdict: FAIL, 104/106 quality gates; both null checks pass

All96 whole-stream/late non-harm gates and6 stationary interaction-gain gates
pass. Both majority-to-parity recovery gates pass. Both parity-to-majority
positive recovery gates fail. Keep these failures despite the near-zero mean
harm and higher pass count. No whole research direction is complete.

The [frozen protocol](mmm-falsification-v99-protocol.md) retains all nine v97
arms and adds a tenth gated arm on768 fresh learned streams. Predictions are
issued before outcomes; a rejected version is masked only on subsequent
forecasts. Fixed refits reopen new versions, not certify them true. Raw expert
predictions and bank learning continue unchanged. A neutral forecast is used
when no unrejected expert has positive weight.

Four experts, eight32-step versions and two views receive a total0.01 test
budget. Equal-weight mixtures over32 starts use threshold6400. Probability
endpoints are rejected by this research contract rather than silently floored.
The monitor's active-sum recurrence is constant-time; it does not scan32 starts
on every update. It matches a literal enumeration in the reference tests.

See [raw learned records and null counts](mmm-falsification-v99.json),
[independent summary with all gates](mmm-falsification-v99-summary.json), and
[evaluator](../../research/falsification-v99-summary.mjs). Learned intervals
are the predeclared paired32-trajectory z=3.5 normal approximation, not exact
coverage or anytime confidence sequences.

## Same-stream confirmation results

| Full-view case and segment | Generic64 Brier | Raw bank | Gated bank |
|---|---:|---:|---:|
| Parity4, all256 | 0.094757 | 0.078360 | 0.078364 |
| Parity4, late128 | 0.070003 | 0.054318 | 0.054318 |
| Majority to parity, late128 | 0.193892 | 0.187827 | 0.180551 |
| Parity to majority, late128 | 0.174502 | 0.178454 | 0.173956 |

Parity4 late expected accuracy is94.91%, and stationary interaction gains
remain above the frozen thresholds. Majority-to-parity late gain is0.013342,
interval[0.007833,0.018850]. Parity-to-majority late gain is only0.000546,
interval[-0.003933,0.005025], failing both the0.005 mean requirement and
positive-lower-bound requirement. Design also fails that recovery gate.

The raw bank on these fresh samples also passes late parity-to-majority
non-harm (upper harm0.006921). Thus we do not claim the gate uniquely rescued
non-harm by comparing it to the different v97 sample. The matched means above
show its contribution, while the full positive recovery claim remains unproven.

Mean full-view masked frames per256-step confirmation stream are10.625 for
majority-to-parity and16.750 for parity-to-majority; neutral fallback means
are0 and0.375 respectively. Stationary parity4 masks0.4375 frames on average
and uses no neutral fallback. These are actions of the detector, not known
false-rejection rates: the learned experts need not satisfy its null.

## Null-model checks

| Mode | Families | Any rejection | Empirical rate | Wilson95% upper |
|---|---:|---:|---:|---:|
| Independent tests | 4096 | 6 | 0.1465% | 0.3192% |
| Shared within version | 4096 | 0 | 0% | 0.0937% |

Both upper bounds are below the declared1% screen. Each independent family
has64 tests. Shared mode has8 version streams copied across4 experts and2
views, preserving64 tests with dependence. This is finite simulated-null
evidence; it is not a universal guarantee that real learned forecasts are
calibrated, or that any rejection signifies poisoning or causality.

Before generation, seed-range review found overlapping draft null namespaces.
They were separated before any null or learned artifact was produced; the
frozen protocol/code contain the final disjoint namespaces. No threshold,
quality criterion or test result was changed to obtain these rates.

## Verification and scope

- Literal start-mixture, neutral/extreme-probability, timing and lifecycle race
  tests PASS:1.237s package time; learned-stream race smoke PASS:3.499s.
- Generation PASS:138.683s; exact learned and null replay PASS:140.686s.
- All17 source hashes, rejection timing/count reconstruction, null intervals,
  summary reproduction, vet and new-file whitespace checks PASS.
- Artifact SHA256:97e2102c8247b4f387eee818bbea5acf158a56b5398db0ce576452c461b37785.

The raw bank's cumulative-loss bound continues to describe only its own raw
output. It does NOT cover the gated forecast or fallback. Sequential testing
error control assumes the target expert's conditional-law null and honest
full feedback. Non-rejection is not a truth certificate. No production or
delayed-feedback path was changed.

## Performance and next lead

Apple M4,Go1.27.1 darwin/arm64,GOMAXPROCS10. The [four-monitor gate kernel](mmm-falsification-v99-kernel-benchmarks.txt)
takes57.79-59.77us for256 predict/update pairs, about226-233ns per pair on
average, with zero allocations across three500ms repetitions. The [full fixture](mmm-falsification-v99-benchmarks.txt)
takes176.75-203.78ms per256-step stream and allocates21.53-21.56MB. It includes
all five constructed models, eight refit rounds, ten arms, two views, scoring
and hashing. Neither figure is per-request serving latency. No cross-run
performance delta against v97 is claimed.

Next test [comparative alternatives](../../research/comparative-falsification-proposal.md)
while keeping the testing budget fixed. Neutrality is a limited benchmark:
a forecast can beat ignorance yet be worse than another available forecast.
Alternative mixtures also dilute evidence and may slow rejection; this tradeoff
must be tested on new data, not asserted to solve the last two failures.
