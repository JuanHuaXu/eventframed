# Frozen query-value critic: negative result

The [protocol](mmm-query-critic-protocol.md) freezes ten features, ridge penalty
.01, phase0-only normalization/training, and both forced and abstaining policies.
It fits4640 query rows with total weight672, one unit per training trajectory.
Phase1 evaluates672 delayed and672 complete-delivery trajectories. Both phases
were previously consumed: this is not untouched confirmation.

## Efficacy

| Policy, phase1 delayed | Actual-answer expected Brier | Queries |
| --- | ---: | ---: |
| No query | 0.168800035 | 0 |
| Random | 0.167334488 | 672 |
| Entropy | 0.166629983 | 672 |
| Original joint | 0.167365847 | 672 |
| Disjoint joint | 0.167744461 | 672 |
| Forced critic | 0.167675322 | 672 |
| Abstaining critic | 0.167534234 | 605 |
| Random with identical abstention mask | 0.167394687 | 605 |
| Entropy with identical abstention mask | 0.166496360 | 605 |

Both frozen advancement screens FAIL. Neither mode establishes all-case
non-inferiority or the required switching-case gains over its matched controls.
Zero of21 cells meet the joint two-control lower-bound non-inferiority screen;
this failure to establish non-inferiority does not prove harm in every cell.
Zero cells establish positive lower-bound gains over both matched controls.

Abstention saves67/672 queries (9.97%) but does not rescue ranking. Stationary
case0 and switching cases19/20 all spend32 queries in both modes. Their critic
Brier is0.217621/0.217098/0.202607 versus entropy0.216107/0.215526/0.201056.
The separate outcome-averaged metric also gives no descriptive advantage over
entropy: forced critic0.167321 versus0.167020, gated0.167340 versus matched
entropy0.167062. No default or publication change is supported.

## Objective diagnostic

A post-hoc decomposition separates within-pool centered target errors from
errors predicting each pool's mean value. The weighted squared-error identity
is checked numerically, and the diagnostic replays byte-for-byte.

| Metric | Phase0 training | Phase1 evaluation |
| --- | ---: | ---: |
| Between-pool mean-value R-squared | 0.2365 | 0.2888 |
| Within-pool relative-value R-squared | -0.0144 | -0.0446 |
| Fraction of target variance within pools | 0.3663 | 0.3569 |

The critic learns some situation-level value but not useful squared-error
prediction of relative candidate value, even in training. Negative within-pool
R-squared alone does not prove its pairwise ordering is always wrong; the direct
policy comparison above supplies the negative efficacy evidence.

A distinct next hypothesis is objective mismatch: fit within-pool centered
features and targets for ranking, while retaining a separate pool-value estimate
for acquisition/abstention. Centering removes between-pool variation from the
ranking loss. This is not yet a confirmed root cause or rescue; limited features,
model misspecification and sampling uncertainty remain alternatives. Do not tune
the present critic on phase1 or erase its failed screen.

## Verification and cost

- Independent pivoted elimination matches the reused Cholesky solution.
  Normal-equation residual is8.67e-19. Degenerate-feature/intercept fixtures,
  ownership, deterministic replay, invalid inputs, ties and abstention pass.
- Feature-access traps reject oracle/identity fields; inference-access traps
  prohibit training targets. Poisoning every phase1 target leaves the fitted
  model and every selection unchanged. Normalization uses phase0 only.
- Selected losses resolve to the earlier audited query branches. Source artifact
  hashes, candidate IDs and equal gated-control query counts are checked.
  This comparison reuses prior posterior fits; it does not refit the Bayesian law.
- Main summary and variance diagnostic each replay byte-identically.
- Offline critic fitting takes3.48-5.06ms over three repetitions. Over2016 warmed
  calls, feature extraction plus selection has median18.42us, p95 24.96us and
  p99 52.96us. First measured selection is168.37us, not a cold-start guarantee.
- These timings exclude constructing both Bayesian decision bundles, retrieval,
  parsing, queues, contention and persistence. The existing oracle supervision
  also costs offline counterfactual fits. No serving-latency or free-supervision
  claim follows. Fast computation does not rescue failed efficacy.

Artifacts in this directory: `mmm-query-critic-v1-summary.json`, its
`-summary-replay.json`, `mmm-query-critic-v1-benchmark.json`, and
`mmm-query-critic-v1-variance.json` with its `-variance-replay.json`.

All seven whole research goals remain open. No production, paper, commit or push.
