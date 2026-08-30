# Output Improvement Ablation Results

## Status

This is a development ablation over 1,286 cases and nine task clusters. The
retired prospective pilot was not reopened. Results select no new accuracy
claim; a later chronological holdout is required for promotion.

The first ablation execution exposed an invalid coupling between lexical rank
and calibrated probability. That execution was discarded. The committed
artifact comes from the corrected contract: `rank_score` orders candidates,
while the proper score evaluates the separately committed probability law.

## Comparison with current EventFrame

| Proposal | Brier change | Recall@10 change | Packed-recall change | Token change | Development decision |
|---|---:|---:|---:|---:|---|
| Contextual Bayesian score | `+0.00041` improvement; interval crosses zero | `-0.00371`; interval crosses zero | `-0.00141`; interval crosses zero | `-1.36` | Unproven; keep experimental |
| Weak hierarchy | `-0.00216`; negative interval | `-0.01334`; negative interval | `-0.01625`, negative interval | `+0.49` | Rejected in this formulation |
| Lexical frontier rerank | effectively unchanged | `-0.00274`; interval crosses zero | `-0.00233`; interval crosses zero | `-0.28` | Unproven; exact-ID unit behavior works |
| Adaptive packing | effectively unchanged | `-0.00030` | `+0.00192`, 95% interval `[0.00071, 0.00319]` | `+9.94` | Supported but low practical value |
| Diversity packing | effectively unchanged | `-0.00030` | `+0.00219`, interval `[-0.00098, 0.00689]` | `+2.14` | Unproven |
| Residual shadow | `-0.000026` | `-0.00030` | `+0.00022`; interval crosses zero | `-0.09` | Applied correction remains negligible |
| Full bundle | `-0.00071`; interval crosses zero | `-0.01501`, negative interval | `-0.01545`; interval crosses zero | `+10.51` | Rejected |

Positive Brier change means lower loss. Packed-recall and token changes use
current EventFrame as the paired control. Cluster intervals resample complete
tasks, not individual turns.

## Absolute reference metrics

Current EventFrame produced Brier `0.15421`, ECE `0.04405`, Recall@10 `0.59569`,
MRR `0.87573`, packed recall `0.45989`, and `1,958.20` estimated tokens per
case. The frozen semantic baseline produced Brier `0.20179`, Recall@10
`0.41597`, and packed recall `0.32232`.

Adaptive packing expanded on `97.28%` of cases but raised mean packed count only
from `7.41` to `7.51` because the token budget was usually already binding. Its
positive packed-recall result is real on this development set, but the trigger
is too broad and the absolute gain is only 0.19 percentage points.

Apple M4 microbenchmarks measured contextual composition at `5.9 ns`, lexical
overlap at `1.78 us`, and combined adaptive/diversity selection over 100
candidates at `1.44 ms` after token sets were cached. The initial diversity
implementation took `4.26 ms` and allocated `9.8 MB`; that implementation was
rejected and replaced before final verification. Embedding latency is not
included in these scoring microbenchmarks.

The hierarchy's weak global parent harmed both ranking and packing despite its
10-percent influence cap. Anti-Pigeon authority remained intact, so this is not
an unsafe merge bug; the broad parent prior is simply a poor predictor for this
task mixture. The full bundle inherits that damage and must not be promoted.

## Structured outcomes

The richer outcome-signal contract is implemented and unit-tested, but these
historical records do not contain explicit citation, rejection, correction, or
test-success signals. It therefore has no retrospective accuracy row. Packed
exposure alone is rejected as learning evidence; rejection and correction
override positive signals.

## Recommendation

Keep current EventFrame scoring. Residual shadow remains available for evidence
collection, but applied residuals retain a tiny development advantage. Retain lexical reranking
and diversity behind flags for future domain-specific holdouts. Do not enable
the hierarchy or full bundle. Adaptive expansion should be redesigned with a
rarer trigger before another frozen experiment rather than tuned against this
dataset.
