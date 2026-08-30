# LibraVDB Contract Retest Results

## Scope

This is a development-only retest over 146 Codex cases from three independent
session clusters. The deterministic slice is capped at four accepted sessions
and one 100-turn segment per session; the chronological 80/20 split therefore
places three clusters in design and leaves the confirmation artifact at zero
cases. This is a bounded contract retest, not a replacement for the full
1,286-case development replay or a later confirmation run.

Every query used LibraVDB 1.9.36-beta.5 through the public
`SearchTextCollections` contract for nomination and `RankCandidates` for final
ordering. Current and future record IDs were excluded per collection before
search. EventFrame resolved every returned ID through its tenant-scoped,
availability-checked event store and preserved the search result metadata when
calling the ranker.

The replay still uses fixed `pack_k` assembly. It does not execute the original
OpenClaw plugin's `BeforeTurnKernel`, exact-recall supplement, or
`AssembleContextInternal`; those remain owned by the original context engine.

## Control comparison

| Variant | Brier | Recall@10 | Recall@50 | Packed recall | Nomination recall |
|---|---:|---:|---:|---:|---:|
| LibraVDB contract baseline | 0.23397 | 41.97% | 91.72% | 28.93% | 99.83% |
| Current EventFrame | 0.18555 | 41.97% | 91.72% | 28.93% | 99.83% |
| Shuffled feedback | 0.25099 | 41.97% | 91.72% | 28.93% | 99.83% |

Current EventFrame improved Brier by 0.04842 over the no-learning baseline.
The cluster-bootstrap 95% interval is [0.03761, 0.05765]. Shuffling feedback
made Brier 0.06543 worse than EventFrame, with interval
[-0.08008, -0.03677]. The probability improvement therefore depends on aligned
feedback in this slice.

All variants had identical Recall@10, Recall@50, MRR, packed recall, and token
use. LibraVDB's ranker did not translate the tested upstream score changes into
a different output order. This retest supports a calibration benefit, not an
output-retrieval benefit.

## EventFrame-relative ablations

| Proposal | Brier gain | 95% cluster interval | Ranking/packing gain | Result |
|---|---:|---:|---:|---|
| Contextual composition | +0.00500 | [-0.01816, 0.01357] | 0 | Unproven |
| Weak hierarchy | -0.00495 | [-0.00638, 0.00227] | 0 | Unproven; adverse point estimate |
| Residual shadow | -0.00012 | [-0.00025, 0] | 0 | Current residual application weakly retained |
| Contextual + hierarchy | +0.00163 | [-0.01215, 0.00862] | 0 | Unproven |

No optional proposal is promoted. The three-cluster intervals are intentionally
wide, and no candidate changed retrieval output under the production ranker.

## Implementation findings

- Local lexical reranking and local adaptive/diversity packing were invalid
  substitutes for methods already owned by the LibraVDB plugin contracts and
  remain removed from the upgrade.
- Search metadata is operational, not decorative. Discarding it and rebuilding
  a synthetic collection name caused rank maintenance to target the wrong
  collection. EventFrame now preserves the exact nomination metadata through
  reranking and fills only missing typed defaults.
- Events outside the contract-nominated frontier are recorded as nomination
  misses with zero retrieval probability; they are not silently reintroduced
  by an exhaustive local scan.
- Every returned frontier member is activated in this experiment. The
  certificates are explicitly conditional on the contract-nominated frontier
  and make no coverage claim about non-nominated corpus events.
- Exact-request caching in the evaluator shares identical nomination or rank
  requests across ablation variants. Distinct score vectors still invoke the
  real rank contract.

## Decision

Keep EventFrame as a bounded predictive overlay between LibraVDB nomination and
ranking. Retain current EventFrame scoring and residual application. Do not
promote contextual composition or hierarchy from this development slice.

Before a production default changes, run the uncapped chronological benchmark,
then test the complete original plugin path including `BeforeTurnKernel`, exact
recall, and `AssembleContextInternal`. The earlier 1.385 ms isolated rank RPC
microbenchmark does not measure maintenance against real stored records and is
not an end-to-end latency claim.
