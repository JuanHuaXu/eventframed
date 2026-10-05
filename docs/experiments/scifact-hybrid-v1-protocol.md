# Public Hybrid Nomination Diagnostic

2026-10-04. Frozen before new predictions. ALL5183 source-pooled EventFrame
documents and the original351FIT queries; calibration180/confirmation300 stay
closed. No outcome labels in the predictor. No production/native daemon edits.

The previous actual-store run verified every document but returned median8
candidates and54.89% query-macro citation recall in the full frontier. Missing
citation targets cannot be rescued by within-frontier ranking. This motivates
an independent lexical nomination experiment, not a diagnosis of the closed
native embedding/ranking internals. Task prefixes/model suffix consumption
remain separate unproven leads.

Implement an immutable inverted index over exactly the stored source-pooled
5W1H text, never annotations or metadata. Lowercase Unicode letter/digit tokens,
no stemming/stopwords; distinct sorted query tokens. Fixed BM25 variant:
IDF=log(1+(N-df+.5)/(df+.5)); score sums IDF*tf*2.2 /
(tf+1.2*(.25+.75*dl/avgdl)). These1.2/.75 constants are design choices, not
fit-selected optima. Zero-match documents are not lexical nominations. Exclude
future documents BEFORE corpus statistics; index immutable at its declared
as-of clock. Equal scores break by canonical record ID. Preserve all source
documents; no corpus reduction or convenient query omission.

Arms: original actual native search; BM25 reranking of EXACTLY that native
frontier (including zero scores); BM25 nomination top200 over the full corpus;
equal-weight reciprocal rank fusion of native top200 and lexical top200, with
fixed k=60 and absent-list contribution0. Fuse the union (at most400), then
truncate to200 BEFORE evaluation/packing10. This is an explicit nomination
change: it may evict native candidates. Record those changes rather than
claiming that union recall survives a200cap. Output all200selected candidates,
not just packed10. Native search is REPLAYED from the fully audited v3 run, not
rerun concurrently; no native rank, Bayesian learning or confidence law change.

Primary diagnostic: frontier citation recall and recall10; MRR10/NDCG10 also
reported. Query-macro and221 source-family-unit macro, paired gains/losses.
FIT comparisons are exploratory and cannot prove untouched agent benefit.
Citation relevance is not support/contradiction, truth or answer usefulness.
Do not select/tune parameters from these results or open confirmation here.

Race tests3times, formula/identity/full-frontier/future/duplicate/corruption/
cancellation controls; independent JS full-corpus BM25 and fusion reconstruction
for EVERY query before reading FIT labels. Measure index construction, postings
and tokens, all serial query/rerank/fusion costs, allocations through benchmarks,
and offline process RSS. Native costs from the prior trace remain separately
reported. Replayed native cost plus new offline cost is a counterfactual sum,
NOT an observed live/loaded p99 or constant-time corpus scalability guarantee.

Research basis: [BEIR](https://arxiv.org/abs/2104.08663) motivates a lexical
baseline across retrieval domains; [Robertson and Zaragoza](https://www.staff.city.ac.uk/~sbrp622/papers/foundations_bm25_review.pdf)
derive BM25; [Cormack et al.](https://plg.uwaterloo.ca/~gvcormac/cormacksigir09-rrf.pdf)
give RRF with frozen60. Our positive IDF, tokenization and finite-list treatment
are explicitly declared implementation choices, not claimed faithful theorem
extensions. Posting traversal scales with query document frequencies; sorting
with matched documents. Future billion-record performance requires measurement.

Patch audit: nomination loss CONFIRMED by full trace/labels; exact upstream
cause NEEDS INVESTIGATION (embedding, filtering, token consumption or rank
contracts). New hybrid technique RECOMMENDATION ONLY, implemented in isolation
under the research goal authorization. Falsifier: no frontier gain or rank harm
on the unchanged fit set; corruption/future leakage is a hard failure, not a
reason to discard queries. Previous native rank/RSS/panic negatives stay failed.
