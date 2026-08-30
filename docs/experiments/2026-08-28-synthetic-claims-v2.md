# Synthetic claims experiment v2

## Scope

This is a mechanism experiment using actual `eventframed` service instances and
the in-memory EventStore. It is not real-world OpenClaw evidence and does not
test the whitepaper's complete marked next-event/no-event law.

The v1 run was discarded for ranking analysis because exact cosine ties exposed
nondeterministic map-order tie resolution. Version 2 adds fixed,
outcome-independent vector perturbations. Two repeated v2 runs produced the
same report SHA-256:
`663eed351a116ab4706c3e77e7d3a07e0048d248a06dc256fc89debbdaa442a4`.

## Frozen design

- 20 memory events in four latent groups, with five events per group.
- 120 chronological predictions across four trajectory identifiers.
- Turns 0-79: the vector-matched group is useful.
- Turns 80-119: a hidden regime shift makes the next group useful while the
  query vector remains unchanged.
- Shift cases receive priority 0.95 before their outcomes are revealed.
- Policy freeze precedes confirmation by a one-hour embargo.
- Every label becomes available one minute after its prediction and before the
  next prediction.
- Common candidate universe, complete recall of 20, and packing of 20.
- Priority weight: `1 + 4p`.
- 2,000 deterministic trajectory-cluster bootstrap samples.

The compared policies are:

1. `baseline`: no Bayesian activation or outcome updates.
2. `update_all`: all 20 nominated events update independent posteriors.
3. `selective`: the declared selective frontier and selected-outcome weighting,
   with residual reuse disabled.
4. `eventframe`: the same selective policy with residual learning and reuse
   enabled.

## Overall results

| Policy | Brier | Priority Brier | Recall@10 | MRR | Activation |
| --- | ---: | ---: | ---: | ---: | ---: |
| Baseline | 0.319899 | 0.343444 | 0.7667 | 0.7133 | 0% |
| Update all | 0.290615 | 0.311532 | 0.7667 | 0.7113 | 100% |
| Selective | 0.319527 | 0.342906 | 0.7667 | 0.7133 | 5% |
| EventFrame | 0.319527 | 0.342906 | 0.7667 | 0.7133 | 5% |

Paired priority-weighted Brier gains against baseline:

| Policy | Absolute gain | Relative reduction | Cluster bootstrap 95% |
| --- | ---: | ---: | ---: |
| Update all | 0.031912 | 9.29% | [0.031838, 0.031977] |
| Selective | 0.000538 | 0.16% | [0.000538, 0.000538] |
| EventFrame | 0.000538 | 0.16% | [0.000538, 0.000538] |

Only four synthetic trajectory clusters exist and their constructions are
nearly symmetric. The narrow bootstrap intervals demonstrate deterministic
paired behavior, not broad population certainty.

## Regime stratification

In the stable period, selective and EventFrame Brier loss are worse than the
baseline by 0.0000046. Update-all improves Brier by 0.02331. All policies have
perfect recall@10 because vector similarity already solves the stable ranking.

After the hidden shift:

| Policy | Brier | Recall@10 | MRR |
| --- | ---: | ---: | ---: |
| Baseline | 0.426922 | 0.3000 | 0.1399 |
| Update all | 0.385690 | 0.3000 | 0.1339 |
| Selective | 0.425797 | 0.3000 | 0.1399 |
| EventFrame | 0.425797 | 0.3000 | 0.1399 |

Update-all substantially improves probability quality but does not repair the
top-10 ranking under this shift. Selective updates lower error only slightly
because the default activation rule activates 5% of the universe. The relevant
shifted group is nominated but generally not activated.

## Claim decisions

- **Residual utility:** not supported in this experiment. `eventframe` is
  numerically identical to `selective`; no incremental residual value appears.
- **Selective Bayesian quality:** weakly supported only for Brier loss overall.
  The 0.16% relative gain is too small to justify deployment on this evidence.
- **Update suppression:** computationally substantial (5% versus 100%
  activation), but it gives up most of update-all's probability improvement.
- **Regime adaptability:** not supported for ranking. No learned policy improves
  recall@10 after the hidden shift.
- **Anti-Pigeon and omitted-influence coverage:** not established. This version
  does not inject divergent posterior-sharing buckets or estimate external
  audit coverage.
- **Real-world usefulness:** untested.

## Reproduction

```sh
go run ./cmd/eventframe-experiment
go run ./cmd/eventframe-experiment -output /tmp/eventframe-synthetic-claims-v2.json
```

Negative results are retained. Thresholds, score weights, or the synthetic
target should not be tuned on this confirmation result; a revised policy needs
a new design block and fresh confirmation experiment.
