# v107 mixture headroom: consumed diagnosis, not a rescue

Status: diagnostic reconstruction PASS. No quality claim is promoted. All seven
research directions remain open. This follows the failed v106 delayed recovery
screen and does not supersede it.

## What was measured

All 128 consumed delayed switch trajectories from v106, across both original
phase labels, and all four post-change publications: 512 model bundles total.
Each bundle was rebuilt from its exact as-of long64/short32 origin lists.
Full-input forecasts were enumerated over the fixture's 512 equally likely
inputs. The oracle minimizes expected binary Brier with one constant convex
weight vector per publication, chosen with hindsight. Four-model and
four-model-plus-neutral optima are separate. Neutral always predicts 0.5.

These are full-input risks, not the served partial-observation Brier. Neither
oracle is available online. This is not a bound on all input-dependent mixtures,
all acquisition policies, or other learners. Lower risk is better; neutral risk
is 0.25 and the true Bernoulli irreducible risk is 0.0475.

## Confirmation-labeled consumed records

Means over 32 trajectories per row. The label identifies the original phase;
these trajectories are now consumed diagnostics, not untouched confirmation.

| Change | Publication | Best pure raw | Best raw mixture | With neutral | Raw mixture worse than neutral |
| --- | ---: | ---: | ---: | ---: | ---: |
| Majority to parity | 128 | 0.285963 | 0.280609 | 0.249958 | 31/32 |
| Majority to parity | 160 | 0.225946 | 0.225364 | 0.221735 | 18/32 |
| Majority to parity | 192 | 0.052149 | 0.051989 | 0.051989 | 0/32 |
| Majority to parity | 224 | 0.048046 | 0.047907 | 0.047887 | 0/32 |
| Parity to majority | 128 | 0.301372 | 0.301341 | 0.248531 | 31/32 |
| Parity to majority | 160 | 0.235351 | 0.232566 | 0.229009 | 7/32 |
| Parity to majority | 192 | 0.102323 | 0.101759 | 0.101716 | 0/32 |
| Parity to majority | 224 | 0.060389 | 0.060297 | 0.060297 | 0/32 |

The design-labeled records agree qualitatively: at publication 128, raw-mixture
means are 0.280315 and 0.290658; adding neutrality gives 0.250000 and 0.248397.
All 16 cells, raw component risks, neutral weights and counts are in the
[machine-readable summary](mmm-mixture-v107-summary.json).

## Interpretation and next lead

Immediately after a switch, selecting among these four stale models is not
enough in most trajectories, even with hindsight and full input. A neutral
fallback materially lowers the best achievable risk within this candidate
class. By publication 192 or 224 the models contain useful forecasts again.
Mixtures can outperform every pure model, so the earlier pure-model comparison
alone was not enough to establish this finding.

This suggests testing explicit neutral competition while retaining the current
evidence gate, separately from origin-aged weighting. The existing neutral
destination is reachable only through evidence-based rejection; it is not a
persistent fifth selector candidate. Do not weaken its statistical threshold
on these consumed trajectories. Do not assume a publication means a real change.
An age-weighted selector is another distinct lead: old revealed losses should
retain their origin age rather than receive renewed influence at late arrival.

Predeclare separate neutral-only, age-only and combined ablations, preserve
unchanged controls, and require stable-case non-harm and both delayed switch
gains on fresh streams. Age discounting may hurt rare valid knowledge; neutral
competition may dilute useful stable confidence. This diagnosis cannot decide
those tradeoffs. No half-life or prior sweep on these outcomes is justified.

Discounted expert prediction is supported as a research direction by
[Chernov and Zhdanov (2010)](https://arxiv.org/abs/1005.1918).
[Joulani, Gyorgy and Szepesvari (2013)](https://proceedings.mlr.press/v28/joulani13.html)
study delayed online learning. Neither source establishes a guarantee for our
particular combination of censoring, changing model roles, acquisition and gates.

## Verification and scope

- Analytic, singular-face, five-way interior and boundary/grid optimizer tests
  passed with the race detector (1.281s); package vet passed.
- All 128 original X/Q/Y streams regenerated; all 1,024 as-of fit-origin lists
  checked, with 512 post-change bundles fitted. Their 16,384 issued generic
  partial forecasts matched exactly. The prior v106 complete policy replay
  remains separate evidence for the other served arms.
- All 28 source hashes verified. Independent JS recomputed quadratics from
  262,144 full-input rows, direct Brier, simplex constraints, first-order gaps,
  nesting and irreducible floors. Maximum numerical convex gap: 2.509e-9.
- Full diagnostic reconstruction replayed exactly: generation 8.52s, replay
  8.53s. These are offline diagnostic runtimes, not request latency or overhead.
- No production path calls the oracle. No runtime, whitepaper, benchmark claim,
  data permissions or external publication changed.

The optimizer's risk-minus-gap is a convex lower bound in exact arithmetic;
Float64 verification is not an interval-arithmetic or statistical certificate.
The [frozen protocol](../../research/mixture-v107-protocol.md) specifies this
limitation. The artifact was created exclusively with mode 0600.

Artifact SHA256:
`d2e2c65b6784695f3a23995dddc66d789b11f5cad56a9df8c065259a814c4c10`.
