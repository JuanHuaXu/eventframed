# Bounded Learned Ranking Diagnostic

2026-10-04, frozen before execution. All5183documents/351FIT queries; unchanged
180calibration/300confirmation remain unused. Existing FIT annotations were
already consumed, so this is NOT untouched validation or an agent success claim.
No native/production edits or new service starts. Native source observations are
the actual fully verified v3 trace, used as a frozen observation snapshot.

BM25 is the stronger control. Use the previous fused200frontier for EVERY arm,
not the shorter native universe. Baseline: BM25 sorting of this frontier; learned:
bounded linear feature correction on the SAME complete200, before packing10.
Compare to original pure BM25 too, to reveal nomination-cap damage. Never call
uncited documents known-false: binary citation judgments are a working relevance
training target, not exhaustive truth or support annotations.

Validate the actual5183 stored rows (exact ID/text/provenance/clock) against the
source-pool registry before exposing a snapshot. Hydrate every nominated row
from that sealed source snapshot; unknown/duplicate/future/tampered rows fail,
no shorter-set fallback. A caller-supplied epoch must equal the raw native trace
SHA256. This is a frozen research snapshot, NOT a live mutation/invalidation
certificate or a claim that reopened sources cannot change in production.

Features contain no IDs, source hashes, family labels or citation annotations.
Pre-tokenize source title and source-pooled5W1H text once. Lowercase Unicode
letters/digits, no stemming/stopwords. Eight[0,1]features: title query overlap,
body overlap, ordered title bigram match fraction, body numeric-token overlap,
log1p(title length)/log1p(8192), log1p(body length)/log1p(8192), native nomination
indicator, and60/(60+native one-based rank), absent0. Clip length ratios at1.
Distinct sorted query terms for overlap; ordered tokens for bigrams. No question
text or outcome label inferred from IDs. Source titles bind to canonical corpus.

Base score=BM25/maxBM25 within the chosen200, or0 when max0. Correction=
.25*tanh(w dot features), independent of beliefs/properly-scored laws. Zero
weights exactly reproduce baseline. Score ties use canonical record ID.
Train pairwise logistic loss log(1+exp(-(s_positive-s_uncited))) plus ridge
.001||w||^2/2. Batch200steps, learning rate .2, weight projection[-4,4]. The
declared tanh derivative is included; no optimizer/hyperparameter search.
Each source-family contributes equally, then its queries equally, then eligible
positive/uncited pairs equally. Queries with no positive in the200 contribute
no pair gradient BUT remain in all prediction/quality denominators. Report
that lack of learning opportunity. No fitting gain is an adoption criterion.

Five folds assign whole FIT source families by SHA256(familyID)'s first32bits
mod5. Each model trains ONLY the other four folds; held-fold labels reach only
the evaluator after predictions. Publish all training query IDs and immutable
weights for exclusion audit. Train an additional all-FIT model for a future
prospective study, not its own validation. No shared-source family leakage.
Retrospective cross-validation is explicitly not a prequential/delayed-evidence
experiment. Previous generator/corpus selection means even cross-fit outcomes
do not establish generalization to untouched tasks.

Research basis: [Burges et al., RankNet](https://www.microsoft.com/en-us/research/wp-content/uploads/2005/08/icml_ranking.pdf)
motivates pairwise logistic ranking. The bounded linear residual, feature map,
family weighting and tanh constraints are our declared research variant, not
the paper's neural model or a guaranteed rescue. [Joachims](https://www.cs.cornell.edu/people/tj/publications/joachims_02c.pdf)
motivates preference learning; citation labels here are NOT clickthrough data.

Race3times, analytic/finite-difference gradient, zero-residual identity,
full-frontier/duplicate/nonfinite/cancellation/concurrency controls. Independent
JS reconstructs all feature values, family folds, training gradients/weights,
and351held-fold rankings before scoring. Measure full feature-index construction,
training and query costs, allocations and processRSS; native costs stay separate.
No inferred live serving/freshness/constant-time corpus guarantee.

Patch gate: prior lexical fit gain CONFIRMED; learned superiority RECOMMENDATION
ONLY. Native filter/token-consumption root causes still NEED INVESTIGATION.
Falsifier: cross-fit recall10/NDCG lose to BM25, or any fold/source/epoch leakage.
The objective remains all seven WHOLE goals; a component pass is not completion.
