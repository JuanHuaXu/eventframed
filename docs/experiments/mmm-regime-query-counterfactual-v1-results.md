# Outcome-averaged query opportunity results

This implements the frozen
[counterfactual protocol](mmm-regime-query-counterfactual-protocol.md).
All 2688 inputs were previously consumed. These are generator-informed,
publication-conditioned action values, not deployed policy measurements.

## Findings

| Action rule | Outcome-averaged expected Brier, delayed runs |
| --- | ---: |
| No query | 0.168932919 |
| Previously selected random query | 0.167342514 |
| Previously selected entropy query | 0.167031063 |
| Previously selected joint query | 0.167493307 |
| Previously selected disjoint-probe query | 0.167432301 |
| Uniform over the candidate pool | 0.167507126 |
| Best expected paid query, oracle | 0.164372674 |
| Best expected choice including abstention, oracle | 0.164215609 |

Each nonredundant query's value averages its two independently fitted outcomes
with the known generator probability. The action rules stay fixed to their
prior predecision selections. Therefore the table differs from their earlier
realized-answer Brier; it does not report a new deployed improvement.

In 1063/1344 delayed runs (79.09%), at least one paid query has positive expected
value relative to no-query. In 219/1344 (16.29%), every paid choice is worse;
the remaining 62 tie within tolerance. Thus useful opportunity persists after
integrating out purchased-answer luck, but universal forced acquisition is not
justified by these data. An abstention mechanism needs its own online evidence.

The best expected paid oracle improves the mean by 0.00265839 relative to the
entropy rule and 0.00313445 relative to uniform-pool acquisition. Across the
42 delayed phase/case cells, descriptive paired mean +/- 3.5 SE intervals for
best-expected-paid gain have positive lower bounds in 24 cells versus no-query,
35 versus random, 24 versus entropy, 32 versus original joint, 36 versus disjoint
joint, and 37 versus uniform. These intervals are not simultaneous or anytime.

Phase 1 cases 19/20 retain expected opportunity: best-paid oracle Brier is
0.207125/0.191106 versus entropy 0.217098/0.200726. The stationary case 0 has
0.215038 versus 0.216576, but five of its 32 trajectories have no beneficial
paid action. Broad stationary protection remains necessary.

The old realized-answer hindsight minimum averages 0.161626333. The difference
from 0.164372674 is descriptive, not a pathwise additive decomposition of luck:
the two minima integrate over different answer realizations.

## Boundaries

The generator uses separate random streams for labels, inputs, delays and
missingness, and Bernoulli label draws conditional on teacher, X and time. The
oracle uses its known Q to average the paid answer and score future predictions.
It also conditions on the fixed natural evidence available at publication161.
Labels already naturally available then are not counterfactually changed;
their paid branch equals no-query. All of that is unavailable, wholly or partly,
to an online choice at160. No Q or future outcome enters fitting or enumeration.

The result rules out "only lucky revealed answers explain all opportunity" on
these consumed inputs. It does not establish that the opportunity is predictable
from admissible predecision features. A next research candidate is a low-capacity
expected-value critic trained offline on synthetic action outcomes, frozen before
evaluation, using only predecision features. It must retain random/entropy
controls, include abstention accounting, and subsequently survive genuinely
untouched tests. Oracle labels or case/seed identities cannot enter its inference.

## Verification

- Six original fixtures pass race, ownership, no-query/redundancy invariance,
  and Q/future-Y poisoning checks; 20 branches are demonstrably non-noop.
- Review caught a coverage gap: that poisoning test was not by itself a clean
  single-label comparison. A separate eight-fixture test now leaves every other
  label unchanged and verifies all37 branch predictions exactly against grouped
  flipping, including a transfer-generator case. Race PASS, package56.120s.
- The supplemental isolation test SHA256 is
  `41f1567f6150bc85f5e489bf92816da8b1a1c2011ecb0ea0e22b6d8b891da8c3`.
  It was added separately without changing the collector's frozen source files.
- All2688 original records exactly match the prior raw envelope artifact.
  Both branches' enumeration, support, cost, redundancy, hash manifests and
  hazard aging pass audit: 1098144 forecast entries, 8999 nonredundant queries.
  These checks do not independently reconstruct every fitted posterior.
- The averaging helper passes112 outcome-pair tests, label-order and boundary
  checks, and six invalid-input cases. Summary replay is byte-identical.
- Collection performs23640 fits in282.12s wall,1118.27s user,5.14s system with
  four workers. This is duplicated counterfactual research work, not hot-path
  latency. The initial race package takes60.648s. `go vet` passes.

Artifacts: `mmm-regime-query-counterfactual-v1.jsonl`,
`mmm-regime-query-counterfactual-v1-summary.json`,
`mmm-regime-query-counterfactual-v1-summary-replay.json`,
`mmm-regime-query-counterfactual-contracts.txt`,
`mmm-regime-query-counterfactual-isolation.txt`, and
`mmm-regime-query-counterfactual-v1-run.txt` in this directory.

All seven whole goals remain open. No production, whitepaper, commit or push.
