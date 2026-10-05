# Contrast-clause focus rescue results

Ran all 44 isolated CaptureTurn/Recall cases: baseline/focus on both public sets.
Local nomic embeddings, memory backend, no generation or learning updates.

| Set | Baseline positive top1 | Focus positive top1 | Interpretation |
| --- | --- | --- | --- |
| Original context | 7/8 | 7/8 | Preserved, no improvement |
| Temporal contrast | 6/10 | 7/10 | One design-set rescue |

Across the two consumed sets, 13/18 becomes14/18 with no loss of a baseline
correct case. This passes the pooled design improvement/non-harm screen, not
independent statistical confirmation. The sets share facts and some phrasings;
18 is a question count, not18 independent domains or trajectories.

Both-members-correct contrast pairs improve1/5 to2/5. The adopted-Pluto query
is rescued by omitting its explicitly contrasted draft clause. Proposed planet
count remains wrong. Pre-decision Pluto and subsequent Ceres remain wrong.
The original ten-of-ten temporal adequacy diagnostic still FAILS.

Only two original-context and four contrast queries change. All unchanged-query
candidate scores and laws match exactly, and all eight records remain present.
Maximum absolute forecast probability movement is0.00977235504231222 on the
original context set and0.010515116406368263 on the contrast set. This is not
a rank-only transformation and gives no calibration or Brier improvement claim.
Absent-answer controls remain unhandled. No latency conclusion is drawn from
these sequential embedding runs.

Artifacts: [context](context-v1/focus-results.json),
[contrast](temporal-contrast-v1/focus-results.json).
Verification: `node research/public-task-pilot/check-contrast-focus.mjs` checks
hashes, all expected cases, exact rewrite, candidate coverage, unchanged-query
invariance, oracle ranks, and law movement. Eight query-transform unit cases
passed with `go test ./cmd/public-contrast-focus`.

## Decision

Do not promote the heuristic. Contrast-clause distraction explains part, not all,
of the failure; deleting an exclusion can discard essential constraints in other
questions. A serious follow-up should represent requested relation, reference
anchor and excluded alternative separately, keeping exclusions for verification
rather than simply deleting them. It needs fresh-domain tests including cases
where the exclusion is essential. That is a research direction, not a completed
extractor, learned representation, or evidence of goal5 completion.

The separate bounded sparse correction also reaches7/10, but rescues a DIFFERENT
case (subsequent Ceres). Combining the methods is not tested here; do not sum
their gains or silently retune them on these results. Runtime and whitepaper
remain unchanged.
