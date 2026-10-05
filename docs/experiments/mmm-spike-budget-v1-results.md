# Pending-feedback budget: pilot passes, broader evidence required

Frozen contract: `mmm-spike-budget-v1-contract.md`. This is a consumed-data
exploratory replay of 84 records / 2,688 forecasts, not untouched confirmation.

| Method | Expected Brier | Realized Brier |
| --- | ---: | ---: |
| Markov incumbent | .178978294 | .186729758 |
| Unguarded feedback mixture | .172047334 | .177179150 |
| Budget-guarded mixture | .172881911 | .178687633 |
| Fixed half-mixture | .174272111 | .179858264 |

The guard retains about 88% of the unguarded average expected-score gain over
Markov. It limits 715/2,688 proposals (26.60%). Forty-four of 84 records improve;
none worsens expected Brier by more than .01. The worst expected-score harm
is .007613163 (phase 1 local-table stationary, immediate schedule). This is
not zero harm, and no confidence interval or population non-inferiority is
established. The earlier unguarded policy had six records harmed above .01.

All four phase/schedule expected-score means improve versus Markov:

| Phase | Schedule | Guard | Markov |
| ---: | --- | ---: | ---: |
| 0 | immediate | .168288484 | .178898229 |
| 0 | delayed | .176734686 | .179315442 |
| 1 | immediate | .168889187 | .176164100 |
| 1 | delayed | .177615286 | .181535407 |

The guard slightly improves on unguarded feedback in the immediate cells but
loses some utility in delayed cells. This is consistent with pending outcomes
occupying the loss budget, not evidence that delay is harmless.

## Verification

The implementation recomputes the arrived/pending ledger before every
forecast. Retrospective auditing sums actual issued losses even for labels
the algorithm never receives, and checks every prefix against the frozen
.01-per-issued-forecast budget. Maximum numerical boundary excess is 8.33e-17
(tolerance 1e-12). Largest cumulative actual excess is .311757455; the terminal
budget is .32. A 512-outcome-sequence small fixture adds 4,608 exhaustive
prefix checks. Current/future/missing-label poisoning, equal-expert,
zero-feedback and reserve-release controls pass. Proposed weights agree
with the independently audited feedback-v1 artifact within 1e-12.

Full replay matches every result and weight exactly, excluding elapsed time.
Postprocessing took .875s and .864s, including source/artifact reads but not
output serialization. This is not serving latency: all 2,101 candidate fits,
2,688 candidate predictions and incumbent costs still apply. The research
ledger is O(32^2) per block; a streaming pending ledger remains unimplemented.
No production changes or new model-race tests were needed for this standalone
research script. The preceding model race checkpoint remains 43.865s PASS.

The pathwise realized-loss bound is supported by the induction in the
contract and these tests. It does not bound conditional expected Brier on
every realized trajectory. The observed expected-score pilot pass is separate
empirical evidence, not a consequence falsely inferred from that bound.

SHA256:

- Script: `f8ce5f6d91b69fea9012a516ee992d1004454adf0ac0f9702b4f2f6d694fab41`
- Contract: `adce2b7b98ee21738cb8536c1f1529ce260cfaea712d175950106300e0f13825`
- Result: `895655f4477f48b9641af2bf243b2575fae749df2a53ad6a69c073173f081ade`
- Replay: `9a93869fb10435b4be09cbbd7a7c7c61e5ba6d7d591721b475f29dce2407c202`

Next: freeze expanded evaluation across indices and clocks using exactly
this prior, candidate, feedback rule and budget. Charge all fits. Include
stationary and shifting scenarios, delayed/missing outcomes, and paired
trajectory uncertainty. Do not tune the .01 allowance on this pilot or count
previously consumed source trajectories as untouched confirmation. All seven
goals remain OPEN; the pilot supports expansion, not deployment.
