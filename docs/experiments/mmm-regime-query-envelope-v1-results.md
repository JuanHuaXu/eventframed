# All-candidate query opportunity results

The frozen [protocol](mmm-regime-query-envelope-protocol.md) evaluates every
available query, not another online selector. All data were previously consumed.
Future generator truth is used only by the offline scorer. These findings
measure hindsight opportunity, not successful acquisition or fresh confirmation.

## Results

| Arm | Delayed mean expected Brier |
| --- | ---: |
| No query | 0.168932919 |
| Random selected query | 0.166979806 |
| Entropy selected query | 0.166656850 |
| Original joint selected query | 0.167327637 |
| Disjoint-probe joint selected query | 0.167300861 |
| Uniform average over the entire pool | 0.167262739 |
| Best paid query in hindsight | 0.161626333 |
| Worst paid query in hindsight | 0.176656761 |
| Best with abstention in hindsight | 0.161609503 |

These averages weight all 1344 delayed trajectories equally. Each paid branch
costs one label. Every trajectory contains the same source environment and
natural arrivals across branches; other counterfactual purchased labels do not
enter that branch's posterior.

In 1314/1344 delayed trajectories (97.77%), some paid choice improves over
no-query. In 18, every paid choice is worse; in the remaining 12, the best
choice ties no-query within the declared numerical tolerance. The best-paid
hindsight mean improves over entropy by 0.00503052 and over the uniform-pool
average by 0.00563641. This disproves the narrow explanation that this pool
simply contains no useful alternatives on these consumed trajectories.

Across 42 delayed phase/case cells, the descriptive mean +/- 3.5 SE intervals
for best-paid hindsight gain have positive lower bounds in 41 cells against
no-query, 29 against random, 22 against entropy, 24 against either joint
selector, and 38 against the uniform-pool mean. These are not simultaneous
or anytime confidence statements, nor claims of an implementable policy.

For phase 1, cases 19 and 20, best-paid hindsight Brier is respectively
0.204946 and 0.189413; entropy is 0.215526 and 0.201056. The stationary case 0
also has hindsight opportunity (0.213008 versus entropy 0.216107).

## Interpretation

The pool is not devoid of useful choices, and the existing fitter can turn
some purchases into better forecasts. However, the hindsight selector knows
both future truth and the realized purchased answer. That is a stronger
information set than any predecision selector. Thus the observed gap can
include irreducible answer luck, model misspecification, and a mismatched
utility target. It does not prove the entire gap can be learned away.

The next discriminating diagnostic should average both possible purchased
outcomes under the generator's known conditional law, with explicit handling
of labels that arrive naturally at publication. This can separate realized
answer luck from expected action value. Such generator-informed calculations
must remain offline oracles, never inputs to an allegedly online policy.

## Verification

- Six fixtures spanning stationary and both switching cases, each delivery
  schedule, pass race, per-branch hidden-outcome exclusion, ownership,
  concurrent replay, and unchanged-policy forecast checks (56.327s package).
- All 2688 records collected successfully: 11965 actual branches, including
  no-query branches; 11820 publication fits including grouped evaluation work.
- The scorer checks 549072 forecast entries including padded branches,
  enumeration, support, costs, redundancy, aging, source/code hashes, and exact
  selected-policy forecast equality against both earlier raw artifacts.
  It does not independently reconstruct each fitted posterior.
- Two synthetic accounting fixtures pass; nine deliberately corrupted
  enumeration/accounting/forecast records are rejected.
- Summary replay is byte-identical. `go vet ./internal/observationlearners`
  and tracked `git diff --check` pass. The latter does not inspect untracked
  research files.
- Collection takes 147.32s wall, 576.93s user, 4.07s system with four workers.
  This enumerates counterfactuals offline; it is not a serving latency result.

Artifacts in this directory: `mmm-regime-query-envelope-v1.jsonl`,
`mmm-regime-query-envelope-v1-summary.json`,
`mmm-regime-query-envelope-v1-summary-replay.json`,
`mmm-regime-query-envelope-contracts.txt`, and
`mmm-regime-query-envelope-v1-run.txt`.

All seven whole research goals remain open. No production, paper, commit,
or push changes are justified by this diagnostic alone.
