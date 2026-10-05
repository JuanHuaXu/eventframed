# Disjoint-probe control results

This is a consumed-data, one-decision diagnostic, not fresh confirmation or
a whole-goal test. The frozen protocol is
[here](mmm-regime-query-disjoint-protocol.md).

## Outcome

Changing only the virtual probe origins from 153..160 to 137..144 removed
origin overlap with the query pool and changed 772 of 1344 delayed selections.
It did not establish a rescue.

| Arm | Delayed mean Brier (lower is better) |
| --- | ---: |
| No query | 0.168932919 |
| Random | 0.166979806 |
| Entropy | 0.166656850 |
| Original joint value | 0.167327637 |
| Disjoint-probe joint value | 0.167300861 |

The new selector improves the descriptive mean over the old selector by only
0.0000267765. Across 42 delayed phase/case cells, the declared paired
mean +/- 3.5 SE intervals have positive lower bounds in three cells against
no query, and zero against random, entropy, or the original selector.
No cell has a negative upper bound against those controls. These finite
diagnostics establish neither equivalence nor a general advantage.

Each paid arm spends 1344 labels. Redundant purchases are respectively
32, 38, 31, and 47 for random, entropy, original joint, and disjoint joint.
The new selected input occupies 0.958% of virtual probe mass versus 0.214%
of future input occurrences. Disjoint origins therefore do not imply matched
input distributions; this result does not exhaust target-distribution design.

## Verification and cost

- Race, ownership, hidden-outcome/future-input leakage, and detached replay
  contracts pass; query-label marginals and entropy choices remain unchanged.
- All 2688 recomputed original records match the archived baseline exactly.
- All three unchanged control arms match exactly within each new record.
- The scorer checks 666624 forecast values across both bundles, support,
  publication aging, selection, costs, and source/code hashes. This is not
  an independent reconstruction of every posterior fit.
- Summary replay is byte-identical; malformed probe-contract tests pass.
- Collection takes 179.50 seconds wall time with four workers, including
  original computation, 1344 additional batches, and 5886 original plus
  5927 candidate publication fits. This is research collection cost, not
  deployed serving latency or a matched runtime benchmark.

Artifacts: `mmm-regime-query-disjoint-v1.jsonl`,
`mmm-regime-query-disjoint-v1-summary.json`,
`mmm-regime-query-disjoint-v1-summary-replay.json`,
`mmm-regime-query-disjoint-contracts.txt`, and
`mmm-regime-query-disjoint-v1-run.txt` in this directory.

All seven goals remain open. No production or whitepaper promotion follows.
