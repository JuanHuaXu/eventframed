# Sparse challenger v1: design cross-validation

Frozen before executing sparse_v1.py. Exploratory: the design fixture has already
been inspected. No result from this run is untouched confirmation.

Use only the eight answer-bearing design queries from design-preflight-v1.json.
Hold out an entire query mission (Mariner10 or Voyager1), fitting on the other
mission's queries. Candidate documents can occur in both folds; this is query
mission transfer, NOT document-disjoint testing. No new retrieval or ingestion.

Features use compressed frames and query only: fraction of query tokens present
in each of six fields; union coverage; bias; and signed, SHA256-hashed indicators
for matched/missing query terms per field (256 buckets, normalized by query term
count). Query tokens capped at 64; query bytes at 1024; each field at 2048 bytes.
Reject oversized data rather than silently truncating. No IDs or labels in
features. Generic stopwords are exactly the earlier v1 adapter's stopwords.

Fit binary logistic regression using 200 full-batch steps, step size 0.5 and L2
0.01 excluding bias. Initial bias logit(0.1), other weights zero. No tuning or
early stopping. Candidate labels mean direct support for this exact task, not
general factual correctness or Bayesian truth. This is an offline challenger,
not yet a port to Go or a streaming learner.

Compare baseline ranking, lexical union coverage, learned logistic score, and
equal-weight baseline/learned score blend. Also report constant 0.1 probability
as a class-prior Brier control. Baseline ranking score and lexical coverage are
heuristics, not calibrated forecasts; their Brier values are diagnostics only.

Report held-out top1, MRR, and per-candidate Brier separately. Stable ties retain
baseline ordering. A useful lead requires at least one additional correct top1
over baseline, no top1 loss against lexical coverage, and lower Brier than the
class-prior control. With eight queries and two clusters these are screening
criteria, not statistical evidence or proof of any roadmap completion.

Preserve all arms and failures. Save predictions and hashes of protocol, script,
raw input, queries and oracle. Model tests must reject invalid inputs and verify
determinism and held-out-label exclusion.

Feature hashing inspiration: Weinberger et al.,
[Feature Hashing for Large Scale Multitask Learning](https://arxiv.org/abs/0902.2206).
No theoretical guarantee from that paper is claimed for this feature map.
