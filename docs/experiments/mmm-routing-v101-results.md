# v101: evidence-attributed routing

## Verdict: FAIL104/106, with both failed means above threshold

All96 whole-stream/late non-harm and6 stationary interaction-gain gates pass.
Both majority-to-parity recovery gates pass. Both parity-to-majority recovery
means now exceed0.005, but their paired confidence lower bounds remain below
zero. The full rescue therefore still fails. All four finite null screens pass.

The [frozen protocol](mmm-routing-v101-protocol.md) adds one routed arm to the
eleven v100 controls on768 fresh streams. Models, raw-bank learning, observations
and comparative rejection tests are unchanged. At first rejection, the new
policy records how much each alternative contributed to the evidence mixture.
Only subsequent forecasts may route the rejected expert's current weight to
neutrality and unrejected alternative recipients. Credits expire with the same
scheduled version reset. Credits are not probabilities of truth.

See [raw records](mmm-routing-v101.json),
[all106 gates and block diagnostics](mmm-routing-v101-summary.json), and
[independent evaluator](../../research/routing-v101-summary.mjs).

## Matched confirmation results

| Full-view scenario and segment | Generic64 Brier | Raw bank | Comparative gate | Routed |
| --- | ---: | ---: | ---: | ---: |
| Parity4, all256 | 0.097637 | 0.077903 | 0.076973 | 0.076972 |
| Parity4, late128 | 0.072337 | 0.053588 | 0.053588 | 0.053588 |
| Majority to parity, late128 | 0.198554 | 0.191988 | 0.176885 | 0.170211 |
| Parity to majority, late128 | 0.172813 | 0.177458 | 0.173363 | 0.167673 |

Parity4 late expected accuracy is95.0%, the generator's noisy-label ceiling.
This is not real-world agent accuracy. Majority-to-parity routed gain over
generic64 is0.028343, interval[0.018526,0.038159].

Parity-to-majority routed gain is0.005140, interval[-0.001183,0.011464].
Design gain is0.007642, interval[-0.000955,0.016239]. Both meet the mean rule,
but neither proves positive recovery under the declared interval rule. Preserve
the FAIL verdict; do not count threshold-sized means as reliable improvement.
These are approximate paired32-trajectory z=3.5 intervals, not exact coverage
or anytime confidence sequences.

The matched confirmation improvement over comparative-only is0.005690 for
parity-to-majority, with descriptive interval[0.002207,0.009173]. That supports
further study of routing, but is not the predeclared generic-control gate and
does not override it. All controls remain in the artifact; no winner was swapped.

## What actually changed

In confirmation parity-to-majority, mean values over each32-step block are:

| Block start | Raw generic32 weight | Routed generic32 weight | Routed neutral weight | Comparative Brier | Routed Brier |
| --- | ---: | ---: | ---: | ---: | ---: |
| 128 | 0.002756 | 0.013286 | 0.213691 | 0.354720 | 0.333725 |
| 160 | 0.015614 | 0.031550 | 0.000010 | 0.211303 | 0.209540 |
| 192 | 0.028056 | 0.028109 | 0.000001 | 0.067064 | 0.067062 |
| 224 | 0.017877 | 0.017894 | <0.000001 | 0.060365 | 0.060365 |

Most of the routed gain is in the first post-change block, before the models
have refitted on new outcomes. Neutrality reduces overconfidence there; routed
and comparative classification accuracy in that block is identical. This is
not evidence that routing already learned the new rule.

At160, generic32 has Brier0.122494 versus generic64's0.222219, but still receives
only3.16% effective mean weight. The useful alternative remains underweighted.
At192 and224 generic32 is worse than generic64, so indiscriminate permanent
promotion remains unsafe. Block diagnostics are consumed explanations, not
oracle inputs to the runtime or new success criteria.

## Null screens and implementation checks

| Null mode | Monitor | Rejected families / total | Wilson95% upper |
| --- | --- | ---: | ---: |
| Independent tests | Neutral | 12/4096 | 0.5114% |
| Independent tests | Comparative | 10/4096 | 0.4489% |
| Shared within version | Neutral | 1/4096 | 0.1382% |
| Shared within version | Comparative | 1/4096 | 0.1382% |

Every upper bound is below1%. Routing inherits no new rejection budget: the
complete comparative test state is checked for exact equality to the control
after EVERY learned-stream outcome. This confirms that routing changed weight
allocation, not which evidence was tested or when models were rejected.
Conditional null checks do not certify fitted forecasts, routed accuracy or
causal claims. Neither routing nor gating inherits the raw bank's loss bound.

- Reference checks cover all16 rejection masks, mass conservation, recipient
  mapping, first-crossing credit freezing, no rejected recipients, fallback,
  invalid inputs, duplicate/wrong outcomes and fixed-version reentry.
- Component race checks pass:1.314s package time; learned race smoke:3.584s.
- Generation passes:142.408s; complete learned/null replay:144.424s.
- Independent summary reproduction checks all21 source hashes, reconstructs
  whole/late scores from block summaries, and checks probability/weight bounds,
  rejection timing and raw-bank loss bounds. Vet and whitespace checks pass.
- Artifact SHA256:1a9f44df7ab2e64211d50b5d20d8842a3cd420bcff259959d8e75980d3e73a38.

## Cost and next step

The [paired arithmetic benchmark](mmm-routing-v101-kernel-benchmarks.txt) on
Apple M4,Go1.27.1 darwin/arm64,GOMAXPROCS10 measures three500ms repetitions:
comparative290.180-293.640us, routing313.077-316.818us per256
forecast/update pairs. That is approximately1.13-1.15us versus1.22-1.24us per
pair, both zero-allocation, roughly8% extra monitor arithmetic in this fixture.
An [earlier benchmark](mmm-routing-v101-kernel-preliminary.txt) is retained:
brief audit commands ran while it was live, so the dedicated repeat above is
reported as the primary run. No lower-cost run was selected by quality outcome.

The [full twelve-arm fixture](mmm-routing-v101-benchmarks.txt) takes177.56-198.32ms
per256-step stream, allocating21.585-21.606MB in2065-2190 allocations. This includes
all model construction, eight refits, both views, scoring and hashes. Neither
the arithmetic nor fixture measurement is per-request production latency.

Next perform [one fixed-size higher-precision replication](../../research/routing-replication-proposal.md)
before adding another mechanism. It must keep the same gates, all controls and
fresh data, with no interim quality look or adaptive sample-size ladder. This
v101 failure remains recorded. All seven research directions remain open; no
production path, private-data source, repository remote or whitepaper was changed.
