# Costed paired audits: better measurement, no policy rescue

**Paired audits reduce estimator RMSE about90% versus the matched two-policy
single-action DR control.** The expected query costs match; realized costs
vary. Confidence intervals remain too wide for promotion and policy gains
remain negative. This does not complete a whole research direction.

[Protocol](mmm-paired-audit-v1-protocol.md),
[experiment](mmm-paired-audit-v1.json),
[independent audit](mmm-paired-audit-v1-audit.json),
[component contracts](mmm-paired-audit-v1-contracts.json).

## Design and scope

Same consumed projection:672 phase0 delayed episodes for regression and672
phase1 episodes for evaluation, repeated under64 logging assignments. These
are not64 new datasets. Both contrasts retain their original query strategies.
The single-action control chooses only between the two compared strategies,
not the earlier three-strategy logger. Exact same-origin choices have known
zero difference and incur zero comparison queries in both designs.

The single design queries one distinct alternative per eligible episode. The
paired design independently includes an eligible episode with probability.5,
queries both alternatives in separate branches when included, and observes no
loss otherwise. Each costs one expected query per eligible episode. Shared
future Y makes paired loss differences less noisy. The branches do not receive
each other's acquired historical labels. Existing source collection calls
fitPopulationQueryBranch separately for each nonredundant origin/answer; this
study consumes those recorded branches, not newly executed production queries.

Both designs need observable future outcomes. The synthetic environment holds
them fixed independently of query choices. This is not a causal estimate of an
interactive agent policy that changes subsequent events. Distinct queries,
separate branch fitting/scoring and ground-truth collection are real prospective
costs; the simulator table must not be presented as free counterfactual evidence.

## Evaluation results

Positive utility is entropy loss minus candidate loss. The fixed-table gains
remain-0.000733161 for random and-0.000857510 for joint8.

| Contrast | Collection | Estimator | RMSE | Mean final CS width |
| --- | --- | --- | ---: | ---: |
| Random / entropy | Single | HT | 0.011509 | 0.207390 |
| Random / entropy | Single | DR | 0.006000 | 0.188705 |
| Random / entropy | Paired | HT | 0.000534 | 0.183119 |
| Random / entropy | Paired | DR | 0.000537 | 0.183119 |
| Joint8 / entropy | Single | HT | 0.012214 | 0.204539 |
| Joint8 / entropy | Single | DR | 0.005136 | 0.187000 |
| Joint8 / entropy | Paired | HT | 0.000524 | 0.183080 |
| Joint8 / entropy | Paired | DR | 0.000516 | 0.183080 |

Paired DR is not consistently better than paired HT. The principal improvement
comes from collecting the paired difference, not from its constant regression.
The unchanged conservative EB bound shrinks much less than RMSE. No one of
the64 assignments yields a positive lower bound in any cell, and none violates
running-average coverage. The family now has eight cells, each alpha=.05/8;
do not compare widths directly to the previous four-cell family as if allocation
were identical. Finite observed coverage is not a proof of the confidence bound.

## Query accounting

| Contrast | Single train / evaluation | Paired mean train / evaluation | Paired evaluation range | Expected train / evaluation, both designs |
| --- | ---: | ---: | ---: | ---: |
| Random / entropy | 576 / 581 | 580.125 / 583.500 | 538-632 | 576 / 581 |
| Joint8 / entropy | 466 / 449 | 465.000 / 449.594 | 400-502 | 466 / 449 |

Counts are nominal paid query requests, not necessarily newly informative labels
if the underlying publication later receives a natural duplicate. Both labels
are charged on every included distinct pair. A hard-cap deployment would need
a separately valid sampling design; this experiment does not claim exact
per-run budget parity. HT and DR reuse their design's observations and must not
be counted as separately purchased datasets.

## Validation

-100 exact inclusion-expectation identities and support checks pass.
-263680 episode expectation checks,20096 same-origin checks and99932 logged
 access/skip checks pass.
-Independent audit reconstructs regressions, coins, costs,344064 prefix
 increments/coverage decisions and512 final interval inversions.
-All point targets match the unchanged fixed table. Full deterministic replay
 is byte-identical: [replay artifact](mmm-paired-audit-v1-replay.json).

## Next decision

Retain paired audits as a candidate evidence-collection design when two isolated
branches can be scored against a common observable outcome. Do not use this
measurement improvement to relabel a failing acquisition policy as successful,
or retune the bound grid until it grants a decision. It also does not justify
retraining the previously failed critic on noisier realized labels: genuinely
new policy/representation work or new task evidence is still needed. All seven
goals remain open. No production, whitepaper, dependency, commit or push changes.
