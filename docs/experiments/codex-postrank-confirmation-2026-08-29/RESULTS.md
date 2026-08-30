# Corrected rank-delta confirmation results

## Result

**Validated on this reserved block.** The corrected EventFrame post-retrieval
ranking passed all five criteria frozen in `PREDECLARED_PROTOCOL.md` before the
confirmation outcomes were scored.

| Frozen gate | Confirmation result | Status |
| --- | ---: | --- |
| Independent trajectories | 3 | Validated |
| Brier gain over LibraVDB baseline | +0.08857; trajectory-cluster 95% CI [+0.04488, +0.13186] | Validated |
| Recall@10 gain | +0.03846; trajectory-cluster 95% CI [+0.01730, +0.13260] | Validated |
| Packed-recall gain | +0.05961; trajectory-cluster 95% CI [+0.02214, +0.15458] | Validated |
| High-priority miss-rate increase | 0.00000 percentage points | Validated |

The confirmation block contains 138 scored cases. EventFrame improved Brier
loss from 0.32094 to 0.23237, Recall@10 from 0.39439 to 0.43285, packed recall
from 0.28564 to 0.34525, Recall@50 from 0.92749 to 0.95840, and MRR from 0.84534
to 0.91207. Mean packed token use decreased by 6.76 tokens per case.

The shuffled-feedback control did not reproduce the aligned result. Its
Recall@10 gain was -0.00572 and its Brier-gain interval crossed zero
([-0.01874, +0.01932]). The update-all frontier control was effectively tied
with EventFrame on this bounded-frontier protocol, as expected after the policy
was changed to update the complete nominated frontier.

## Contract correction encountered before scoring

The first full attempt stopped in the design block before writing any result
artifact because LibraVDB's second-pass `RankCandidates` contract returned a
scored subset of the nominated frontier. Live contract probes reproduced this
as deterministic low-score-tail omission; they falsified a generic size cap and
duplicate-ID explanation. EventFrame now validates every returned item and
retains unreturned nominated candidates at their frozen base score, allowing
post-retrieval deltas to rescue them. The sealed confirmation block remained
unscored until this invariant had a regression test and the clean run above was
started from a new sidecar database.

## Limits

This is a small three-trajectory chronological confirmation, so the clustered
intervals are broad. The labels measure downstream retrieval usefulness, not
causal task success or the whitepaper's full marked-event law. The result does
not validate 5W1H superiority, Anti-Pigeon coverage, omitted-influence coverage,
or large-corpus tail latency.
