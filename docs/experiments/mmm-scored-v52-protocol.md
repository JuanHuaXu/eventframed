# All-mixture score alignment V52

## Classification and Boundary

Recommendation only, not a confirmed delayed-context bug. V51 already owns
original advice and replays delayed labels at their original positions. Its
four modes change ONLY the outer head. Three credible reasons for its weak
results are: auxiliary objective mismatch, poor child advice, and insufficient
task-local evidence. This trial tests the first; a failure leaves the others
open. No production, existing sources, private data or held-out agent labels
are changed. Old negative results remain immutable.

## Frozen Candidate

Keep Full, Adaptive and memoized rich-moment2 child learners unchanged. Use
the already audited scoreModelV51 at member-local, global and outer heads.
Each mode is uniform at all mixture heads: log_mean, log_strong, brier_mean,
brier_strong. Priors (.8,.1,.1), (.9,.1), member hazard 1/16, global hazard
1/2400, scope hazards static/slow/round, member cap16, global2400 and all
three arrival schedules retain the original contracts. Nested score weights
are strategies, not independent Bayesian copies of one outcome.

The log_mean negative control must match V49 legacy non-cost outputs exactly.
Every child's as-of advice must match the original control for every mode;
local/global head advice and weights are expected to differ in other modes.
Every auxiliary and served forecast must match an independent dense reference
using original-position evidence, next-position transitions, and independent
positive-part substitution. Original served receipts, owner, epoch, cancellation,
caps and whole-bundle fencing remain required. Future-label forks must preserve
all pre-reveal forecasts/receipts/snapshots and alter the final law. Corruption
controls must reject changed advice, weights, receipts, law, cost and coverage.

## Evaluation

Fresh diagnostic seed base2026105207; reserve design2026105209 and confirmation
2026105211 with16replicates per cell, only after a promising screen. Verify actual
generated seed separation, not just base inequality. Diagnostic includes ALL
14regimes x2geometries x3arrival schedules x3scope hazards x4modes and same-world
legacy; 28worlds, 67200distinct outcomes. No favorable-cell pruning. Report
ordinary and priority expected Brier, final top10 usefulness and recovery,
against both Full and Adaptive and legacy. Diagnostic n1 cannot establish
uncertainty, generalization, valid error control or close any of seven goals.

Keep existing future-normal requirements unchanged: Full improvement >=.01
with positive lower confidence bound; Adaptive harm bounded by.01; shifted
recovery improvement >=10% with positive lower bound; complete-loop400ms
and constructor8MiB. Time ALL construction, issue, resolve and snapshots,
not just the tiny score update. These are offline loop gates, not loaded100ms
serving/RSS/freshness proofs. Report all modes even when a gate fails.

## Research Sources

[Vovk and Zhdanov, Prediction With Expert Advice For The Brier Game (2009)](https://jmlr.org/papers/v10/vovk09a.html)
supports Brier mixability and substitution, not this nested delayed design.
[Joulani et al., Online Learning under Delayed Feedback (2013)](https://proceedings.mlr.press/v28/joulani13.html)
analyzes delay and black-box transformations; we do not claim its regret
guarantees for retained-position replay. Here the theorem-free falsifier is
the full controlled experiment: alignment fails if useful quality/recovery
does not improve or cost/stationary harm violates the original gates.
