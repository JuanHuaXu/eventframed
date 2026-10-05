# Centered query ranking: partial fit improvement, failed screen

The [frozen objective control](mmm-query-centered-critic-protocol.md) changes
within-pool training, not the ten features, penalty, original mean-value head,
data split, or acceptance thresholds. All data remain previously consumed.

## Results

| Phase1 delayed policy | Actual-answer expected Brier | Queries |
| --- | ---: | ---: |
| No query | 0.168800035 | 0 |
| Random | 0.167334488 | 672 |
| Entropy | 0.166629983 | 672 |
| Previous forced critic | 0.167675322 | 672 |
| Centered forced critic | 0.167148793 | 672 |
| Centered gated critic | 0.167068293 | 588 |
| Random with centered critic's mask | 0.167434543 | 588 |
| Entropy with centered critic's mask | 0.166443884 | 588 |

Both advancement screens FAIL. Forced mode meets the joint two-control
non-inferiority lower bound in0/21 cells; gated mode in2/21. Neither establishes
the required gains on switching cases19/20, and neither has a cell with positive
lower gain bounds versus both matched controls. Failure to establish
non-inferiority is not evidence of harm in every cell.

The forced critic improves its descriptive actual-answer mean over its prior
version by0.000526529, but remains0.000518810 worse than entropy. Its separate
outcome-averaged mean is0.167323883, effectively unchanged from the earlier
0.167320737, and worse than entropy0.167019776. These metrics do not validate
faster learning. Gating saves84/672 queries (12.5%), but matched entropy remains
better in the descriptive actual-answer mean by0.000624410.

## What changed

Within-pool R-squared improves from-0.01435 to0.01390 in training and from
-0.04457 to0.00875 in evaluation. Between-pool R-squared is exactly unchanged,
as required by preserving the original mean-value head. The remaining explained
within-pool variance is below1% on evaluation. Objective mismatch explains some
fit error, but changing this objective alone does not establish a useful rescue.

The current ten features provide uncertainty, hypothetical gain and support-age
summaries. They do not include observed prediction-error information or rich
learner-state interactions. A [primary-source follow-up](mmm-query-critic-literature-followup.md)
identifies a distinct lead; no penalty/feature sweep on the failed phase1 screen
is authorized by this result.

## Verification and cost

- Context-head coefficients exactly match the old critic; all old control
  selections, losses and costs match the prior artifact.
- Pool-mean preservation, common feature/target shift invariance, inference
  oracle-field traps, degenerate pools, detached replay and invalid inputs pass.
- Audit identified a target-domain edge case before collection: centered gains
  can span[-2,2] while the reused fitter accepts[-1,1]. Reversible factor-two
  scaling preserves the linear ridge solution without clipping. An extreme-gain
  regression fixture passes.
- Poisoned phase1 targets leave both fitted heads and all choices unchanged.
  Training uses4640 rows over672 phase0 delayed pools only.
- Summary and variance-diagnostic replays are byte-identical. The rank-head
  normal-equation residual is5.42e-20. All84 phase/case/schedule cells are retained.
- Fitting both heads takes6.74-13.32ms over three repetitions. Across2016 warmed
  calls, feature extraction plus selection has median22.83us, p95 29.83us,
  p99 59.83us; first measured call231.79us. This excludes Bayesian decision-bundle
  creation, retrieval, parsing, queues, persistence and contention. It is not
  a serving benchmark or a reason to deploy a failed efficacy candidate.

Artifacts: `mmm-query-centered-critic-v1-summary.json`, its summary replay,
`mmm-query-centered-critic-v1-variance.json`, its variance replay, and
`mmm-query-centered-critic-v1-benchmark.json` in this directory.

All seven whole research goals remain open. No production, paper, commit or push.
