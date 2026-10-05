# Unrestricted segment mixing ablation

Status: FAIL as a broad rescue. This is a consumed-data screen on 672 schedule
runs with eight indices per cell, not the original 32-index confirmation.
[Contract](mmm-segment-unrestricted-v1-contract.md) was written before scoring.
All seven research goals remain open; no production or runtime promotion.

## Results

The only change from the guarded family is removing its per-outcome restriction.
Two existing heads are mixed with the same Markov baseline: adaptive delayed
Fixed Share, or fixed half as a matched non-learning control. No new fits,
retention changes, parameter grid, or oracle inputs enter the candidate.

| Arm | Whole Brier | Terminal64 Brier | Protection / 672 | Gains / 96 | Harm windows / 5376 |
| --- | ---: | ---: | ---: | ---: | ---: |
| Segment adaptive | .156123153 | .144098212 | 606 | 7 | 172 |
| Static adaptive | .157529244 | .145426054 | 611 | 1 | 148 |
| Segment half | .156590304 | .144586055 | 592 | 8 | 202 |
| Static half | .157987032 | .145546397 | 603 | 3 | 218 |

Markov whole Brier is .157516545. Segment adaptive improves the pooled mean by
.001393392, approximately .885% relative. Its changing-scenario whole Brier is
.214946589 versus Markov .218930662, but stationary Brier is .119924116 versus
.119723242. There is still a small stationary mean regression.

The screen keeps the original .01 upper non-harm tolerance and .005 mean gain
with positive lower endpoint. Intervals are approximate mean +/-3.5SE using
the eight available indices, not simultaneous or anytime guarantees. Counts
are descriptive and cannot substitute for original sample size or fresh data.
Short-window harms are individual trajectory diagnostics, not additional
retroactive acceptance gates.

## Where the failures are

For segment adaptive, non-harm passes against generic64 are 159/168, Boolean64
133/168, Markov 162/168, and static64 152/168. Against Markov, the six failed
screens are:

| Phase:case:schedule | Period | Mean excess | Upper endpoint |
| --- | --- | ---: | ---: |
| 0:2:0 | terminal64 | -.000556121 | .012015531 |
| 0:19:1 | whole | -.008557271 | .010723398 |
| 0:19:1 | terminal64 | -.009435449 | .025710703 |
| 1:19:0 | terminal64 | -.008612984 | .019177473 |
| 1:19:1 | whole | -.006067543 | .014459388 |
| 1:19:1 | terminal64 | +.001876494 | .031661554 |

Case2 is additive gradual; case19 is majority-to-parity. Five failed screens
have favorable means but insufficient precision to establish non-harm. The
sixth has a small adverse mean and substantial uncertainty. Therefore these
six failures must not all be described as demonstrated positive harm. They
also do not license replacing the uncertainty test with a point estimate.

Removing the guard releases useful headroom but does not meet the full
comparison contract. Adaptive mixing gains more on average than the fixed-half
control, yet has only 7/96 qualifying recovery comparisons. Neither label
availability nor the choice of guard has been proven the sole remaining cause.

## Checks and timing

- Existing 384 literal-path delayed-filter checks pass (maximum error 3.89e-16).
- Current/future/unavailable-label poisoning leaves all prior forecasts intact;
  equal forecast inputs remain unchanged. Only the explicit b/c/y/delay/missing
  tape enters mixing; generator Q is used after forecasts for evaluation.
- Every raw control's whole and terminal Brier agrees with source metrics
  within 1e-12. All 672 identities and eight indices per cell are verified.
- Full screen replay is byte-identical. Timing is stored separately.
- One-pass mixing takes 559.78ms, about 1.63 microseconds per head forecast,
  including adaptive and half outputs. This includes cold/JIT effects and
  excludes input preparation, I/O, scoring, model fits, and loaded serving.
  It is not a stable microbenchmark or a daemon tail-latency result.
- Source model cost remains the original 545.32s collection, not zero because
  this ablation reuses stored forecasts.

## Research-source applicability

[Huang, Ma and Michailidis (2026)](https://proceedings.mlr.press/v337/huang26b.html)
was read through sections 3-4, including equations 17-26. It uses a gated
Bayesian residual head, source replay anchoring and antecedent-input mismatch.
Crucially, its computable objective is distinguished from the full certificate:
the latter includes a residual mismatch requiring target labels, and the
forecasting proxy is not directly a raw-error bound. It therefore does not
provide a drop-in guarantee for our delayed binary-Brier correction. Input
disagreement alone cannot identify arbitrary changes in the label mechanism.
No implementation or theorem from that paper is claimed in this ablation.

Downloaded primary PDF SHA256:
`cb56f44a3dcf270dc23841f619ff50509553bb3e6ee2512740a969c200659e70`.
It is scratch research material, not a new distributed dataset.

## Next work and artifacts

Before another weighting change, separate uncertainty from systematic error
in the six Markov failures and inspect the much larger Boolean-control gap.
Any expanded sample or new candidate needs a frozen comparison; do not keep
adding samples until intervals happen to pass. Existing interval-surrogate,
Squint, hazard-mixture and static-atom failures remain relevant negatives.

`research/segment-unrestricted-screen.mjs` produces the screen and separate
timing artifacts from the existing independent source. It refuses existing
output paths. Source SHA256:
`44dd9e651aa93d50753467bf7d29780255e76504bb4d8ee62ba5e328583153df`.
Screen SHA256:
`affe7f545f3dbaa4783132061187db5ef08399b4ddf887ad12699241567eb906`.

No Go, production, private-data, whitepaper, commit or push changes.
