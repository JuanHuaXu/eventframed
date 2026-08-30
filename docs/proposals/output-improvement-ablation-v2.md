# Contract-Native Output Improvement Ablation v2

## Status

This document supersedes `output-improvement-ablation-v1.md`. The v1 lexical,
adaptive-packing, and diversity-packing comparisons are invalid for production
architecture decisions because they substituted local heuristics for retrieval
methods already owned by the LibraVDB OpenClaw contract.

This remains a development experiment. No policy may be promoted from the
consumed development data or the retired prospective pilot. Promotion requires
a later chronological holdout with at least three independent task clusters.

## Frozen ownership boundary

The production retrieval order is:

1. LibraVDB `SearchText` or `SearchTextCollections` nominates candidates and
   supplies the initial hybrid retrieval score.
2. EventFrame applies an optional certified Bayesian/Anti-Pigeon adjustment to
   that score. It does not scan the corpus or invent a competing lexical index.
3. LibraVDB `RankCandidates(k1,k2)` performs the declared two-pass rerank using
   candidate text and complete backend metadata.
4. The original context engine adds `BeforeTurnKernel` predictions and bounded
   exact recall where applicable.
5. `AssembleContextInternal` owns token-budgeted context assembly.

Every variant uses the same LibraVDB contract instance and metadata mapping.
Backend ranking may change order but cannot change EventFrame's forecast law.
EventFrame probabilities may change only through a certified posterior or an
eligible residual.

## P1: Contextual Bayesian overlay

For a candidate with initial hybrid score `s` and certified posterior mean `p`,
EventFrame supplies the bounded pre-rerank score

```text
0.85 * s + 0.10 * p + 0.05 * recency
```

When no certified posterior is available, posterior weight returns to `s`.
LibraVDB then receives this value as `RankCandidate.score`; its two-pass ranker
retains final ordering authority.

## P2: Weak hierarchical prior

Maintain an event or Anti-Pigeon-group Beta posterior and a tenant/horizon Beta
parent. The child borrows at most two effective observations and parent
influence is capped at 10 percent. The parent cannot create or merge an
Anti-Pigeon bucket, bypass a changepoint, or alter LibraVDB metadata.

## P3: Residual shadow

Evaluate residual eligibility and record outcomes without changing the returned
law or pre-rerank score. Compare shadow mode with current residual application
under the same contract-native ranking output.

## P4: Structured outcome evidence

Outcome requests may declare non-content signals: packed, cited, correction,
successful downstream use, user rejection, and explicit usefulness. Rejection
or correction is negative; citation or successful downstream use is positive;
packed-only exposure is not evidence. Free-form transcript content is not
stored in posterior records.

## Removed duplicate proposals

- Local lexical reranking is removed. Exact and hybrid retrieval already belong
  to LibraVDB and the original context engine.
- Adaptive local packing is removed from the production experiment.
  `AssembleContextInternal` owns final token-budget enforcement.
- Local diversity packing is removed. Any future diversity contract must be
  added to LibraVDB assembly and evaluated there, not silently layered after it.

## Ablations and controls

Compare contract-native baseline, update-all without EventFrame residuals,
current EventFrame, contextual-only, hierarchy-only, residual-shadow,
contextual-plus-hierarchy, and shuffled-feedback controls. Primary metrics are
Brier, ECE, Recall@10, MRR, packed recall, high-priority miss rate, token use,
and contract RPC latency.

The development runner must use the real
`libravdb.ipc.v1.LibravDB/SearchTextCollections` nomination RPC and preserve its
result metadata through the real `RankCandidates` RPC. A pass-through ranker or
embedded search is allowed for unit tests and degraded operation but cannot
support a production retrieval claim. Contract-nominated misses remain misses;
the evaluator must not silently restore them with an exhaustive local scan.
Confirmation data remains untouched during proposal revision.

The bounded 2026-08-29 contract retest found that all score variants produced
identical retrieval order and packing on its 146-case, three-cluster slice.
Contextual composition improved the Brier point estimate by 0.00500 relative to
current EventFrame, but its cluster interval crossed zero and no output metric
changed. No proposal is promoted from that retest.
