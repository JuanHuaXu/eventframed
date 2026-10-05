# Real-field research adapter v1

Declared2026-09-12 before fitting/evaluation on the public task outcomes.
Package `internal/researchmemory`, disabled/unwired in serving. This is a
testable transfer hypothesis, not a proven semantic representation.

Input: query<=1024 bytes, six UTF-8 5W1H field values<=2048 bytes each, and an
already-journalable baseline usefulness probability. Content, IDs, attributes,
oracle labels and future feedback are excluded. Nine binary features:

1. Query token overlap with Who.
2. Overlap with What.
3. Overlap with Where.
4. Overlap with When.
5. Overlap with Why.
6. Overlap with How.
7. At least half of non-stopword query tokens appear in the six-field union.
8. Query has numeric tokens and all appear in that union.
9. Baseline probability>=.5.

Lowercase Unicode letters/digits form tokens; frozen stopwords are in source.
Numbers are matched lexically, not parsed into entity roles or dates. This can
conflate a mission number and an unrelated date component. Lexical overlap does
not imply entailment, shared cause, trustworthiness or calibrated confidence.
Every feature is observed; no uniform/independent missing-bit model is assumed.

Retained learner uses latest64 count and tree learners plus long256 counts,
with v8's adaptive inner short/tree mix and preserved external baseline. Fits
begin at32 labels, then every16; before support the baseline passes unchanged.
Weights update only from journaled ready pre-outcome experts. This warm-up is
an explicit transfer departure from the synthetic pilot, not silent equivalence.
The baseline is externally supplied per candidate, not a synthesized true prior.

Each candidate prediction gets a monotonically assigned local ID, epoch and
timestamp. Pending records cap256; saturation rejects rather than overwriting
unobserved evidence. Explicit labels must arrive after prediction and in
nondecreasing availability order. Repeats/stale epochs reject. Discarding a
record adds no negative label. Whole new epoch uses a fresh adapter. No automatic
self-labeling or shared-posterior authority is added.

All count/tree features and training buffers are bounded. Current refits are
synchronous inside a mutex: this API is offline/slow-path, not the finished
immutable-snapshot production adapter. Scope excludes persistent feedback,
source-dependence weighting, AP certification and deployment. Integration must
not feed the same evidence through this learner and a Bayesian path twice.

Public-pilot evaluations must score every candidate forecast before releasing
any query's relevance labels, freeze confirmation learning, and compare with
both unchanged baseline and shuffled-design labels. Candidate pairs from the
same query are correlated; they are not independent test samples. A positive
small result would license more tasks, not establish real-world generalization.
