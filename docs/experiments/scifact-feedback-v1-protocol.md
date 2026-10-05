# Frozen Source Feedback Development Screen

2026-10-04, before predictions. All5183actual source-bound framed documents/
all351CONSUMED SciFact FIT queries/221source-overlap units. No NFtest or CAL
tuning, no CONFprediction. Static isolated nomination research, NOT continuous
learning, live store freshness or observed answer usefulness. Native verified
closed-copy provenance is reused read-only; no production/native service started.

Hypothesis NEEDS INVESTIGATION: failed NFtransfer is largely nomination-limited;
bounded source-derived expansion might improve nomination without losing useful
top10order. Competing causes: topic drift from incorrect top sources, incompatible
fixed rank residual, lexical/semantic gap unrepairable by feedback. No claimed
root cause or acceptance of pseudo-relevance as truth. Falsifier: primary loses
recall10/NDCG10 versus stronger controls or loses frontier recall. Prior native
fallback, rank-weighting and calibrated-transfer negatives remain preserved.

RM3-inspired (not bit-identical Anserini): current Unicode letter/digit tokenizer,
unique-original-query uniform mass; original BM25k1=1.2,b=.75. First BM25top200,
top10feedback documents. Perdoc feedback terms:2..20Unicode scalars anddf/N<=.1,
take top10byTF with lexical ties, L1normalize retained counts; empty feedback
doc contributes zero. Sum doc term vectors weighted by positive BM25scores,
global top10weighted terms, L1normalize; mix original query mass.5/feedback.5.
If no feedback mass, use full original query mass. No orphan query padding,
word-level truncation before corpus scoring, corpus/label-conditioned tuning or
feedback-state writes. All postings of <=266mixed terms scored, THEN top200.
Term mass is a search weighting, NOT calibrated truth/relevance probability.

Preserve ALL queries, including empty nominations, and ALL nominees before
pack10. Four arms: pure BM25, pureBM25+existing held-fold sourceIDF correction,
feedback-weightedBM25, feedback+SAME held-fold correction. Models from prior
IDFstudy five family folds, exact8finite bounded weights; no refitting.
New nomination does not inherit old residual certificate or claim a probability
law benefit; this bounded offline ranker remains search-order-only.

Primary development screen: feedback+IDF must exceed BOTH plainBM25 and
BM25+IDF controls in recall10/NDCG10 in query and221equal-unit means, with
feedback frontier recall not below plainBM25 in bothaggregations. Screen ONLY,
not adoption/untouched validation. Use ALL fixed folds/families/cases; no
query bootstrap, source splitting, parameter selection from consumed NF/CAL.
Different truly unforecasted cohort required after freeze if promising.

Freeze source/test closure/protocol/receipt/model inputs. Predictor no labels
argument; family is model-selection index only and not a scoring feature.
Independent evaluator rebuilds full term/posting snapshot, baseline ranking,
feedback weights and every four-arm full ranking BEFORE consuming FITlabels.
Race/vet/build/full experiment, source/epoch/model identity and deliberate
feature/query/nomination/feedback/future corruption checks. Count initial and
expanded full-corpus scoring, duplicated maps, hydration, features, ranking,
trace writes/fsync/process RSS and source-index build. No cached-rank-only or
unmatched native-latency speedup claim; no loaded100msgoal6completion claim.

Sources: [Lavrenko & Croft2001](https://ciir.cs.umass.edu/pubfiles/ir-225.pdf),
[official Anserini RM3 implementation](https://github.com/castorini/anserini/blob/master/src/main/java/io/anserini/rerank/lib/Rm3Reranker.java).
Borrow estimation/interpolation idea; Unicode/unique-query conventions and fixed
BM25 constants remain this specification. No upstream source copied/licensing
change or toolkit installed. New module clones our existing research-only BM25
core; original files frozen/unchanged, no production/shared package refactor.
