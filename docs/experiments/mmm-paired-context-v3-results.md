# Paired current-context certificate v3: complete-pair component pass, sparse failure

The [frozen protocol](mmm-paired-context-v3-protocol.md) **fails its full
transfer screen** on fresh confirmation seeds. Contemporaneous, same-context
reference/live pairs distinguish a member-only conditional shift from common
co-drift when every pair is available. Under 25% nomination, 20% independent
per-member missingness, and 0..31-clock delays, the same fixed certificate
never flags a shift in 1,000 trials per shift case. This is a component
result, not actual MMM Anti-Pigeon authority or a downstream forecast win.

## Method and reproducibility

The candidate tests each of two context cells over the last 128 origin clocks,
using paired differences `D=Y_ref-Y_live` and a fixed finite-horizon
Hoeffding radius. A common shift changes both members' outcome laws but keeps
the conditional mean of `D` at zero. The comparator gates use the same
delivered pairs: cumulative conditional means and a context-blind 128-clock
mean. Nomination and missingness are outcome-independent. A pair enters a
gate only after both labels arrive; a late pair outside the origin window is
ignored. The isolated simulator rejects future, unnominated, missing, and
duplicate pair deliveries in its negative controls.

- [Simulator](../../research/paired-context-v3.mjs), [independent verifier](../../research/paired-context-v3-verify.mjs).
- [Design](mmm-paired-context-v3-design.jsonl): 10,000 trials; SHA256 `3fb147e40287306623fb5df8de8dde7bb04b80e376c99ebbf5014b5ea407d372`.
- [Confirmation](mmm-paired-context-v3-confirmation.jsonl): 10,000 fresh trials; SHA256 `857d6874c7cc8a76ba49c4d81e8ba94af84adcc9a5f9fe68d46e6f91b9c73f38`.
- Both artifacts pass the independent verifier's source/protocol hash,
  row-key, count, resource, and summary checks. A fresh confirmation replay
  matches all 10,002 JSONL entries exactly after excluding elapsed time.

## Confirmation result

Each entry is flagged trajectories / 1,000. The median clock is conditional
on a flag and is not a population recovery-time estimate.

| Schedule and case | Paired window | Cumulative | Context-blind window | Candidate median clock |
| --- | ---: | ---: | ---: | ---: |
| Complete, stable | 0 | 0 | 0 | none |
| Complete, live-only shift at 256 | 1,000 | 821 | 0 | 361 |
| Complete, reference-only shift at 256 | 1,000 | 786 | 0 | 361 |
| Complete, common shift at 256 | 0 | 0 | 0 | none |
| Complete, live shift from start | 1,000 | 1,000 | 0 | 82 |
| Sparse-delayed, stable | 0 | 0 | 0 | none |
| Sparse-delayed, live-only shift at 256 | **0** | 0 | 0 | none |
| Sparse-delayed, reference-only shift at 256 | **0** | 0 | 0 | none |
| Sparse-delayed, common shift at 256 | 0 | 0 | 0 | none |
| Sparse-delayed, live shift from start | **0** | 411 | 0 | none |

No member-specific shift flagged before clock 256. Design results agree
qualitatively: both complete member-only cases flag 1,000/1,000 at median
clock 361, common co-drift flags 0/1,000, and every sparse-delayed candidate
shift case flags 0/1,000. The fixed full-study pass condition is false in
both splits. The zero false-flag counts are empirical observations, not proof
that target-law or multi-bucket error is zero. For one 0/1,000 cell, the
two-sided 95% Wilson upper endpoint is about 0.383%, without a simultaneous
adjustment across cells.

Complete acquisition costs 512 pairs, 1,024 context readings and 1,024
observed/usable labels per trajectory. Sparse-delayed confirmation averages
about 128 nominations, 256 context readings, 199 observed labels, but only
79 usable pairs (158 usable labels) over all 512 clocks; roughly three
nominated non-missing pairs remain pending at the horizon. The mean values
vary slightly by case and are recorded per cell in the JSONL summaries.
An observed unpaired label is counted as acquisition cost even though it
cannot update this gate.

The sparse failure has a structural explanation. Since `|mean D| <= 1`,
the fixed threshold `epsilon+r_cond(n)` cannot be exceeded unless a context
cell has at least 29 pairs in its active 128-clock window. The sparse
schedule yields about 79 pairs over the *entire* 512 clocks, or roughly ten
per cell in a typical active window. Thus the certificate is effectively
silent at this audit rate; lowering a threshold after seeing confirmation
would forfeit the frozen error statement.

## Scope and next decision

This study repairs the v2 stationary-reference assumption only under a
stronger observation contract: contemporaneous matched contexts and
independent, outcome-blind pair acquisition. Ordinary asynchronous
EventFrame streams may not supply those pairs. The reported ~2.2 seconds
per 10,000-trial split is experiment wall time, not service latency.
There is no scored-law, answer-quality, multi-bucket, or loaded-serving test.

Keep production unchanged and Goal 3 open. A successor should first solve
the acquisition-power tradeoff: compare a variance-adaptive anytime-valid
paired statistic and a predeclared adaptive pair-allocation policy at the
same *total* observed-label/context cost. It must retain a common-drift
negative control and test actual match availability; an unmatched event
cannot be silently treated as a contemporaneous pair. The classical
bounded-difference tool used here is [Hoeffding (1963)](https://www.tandfonline.com/doi/abs/10.1080/01621459.1963.10500830);
no stronger theorem is claimed for MMM.
