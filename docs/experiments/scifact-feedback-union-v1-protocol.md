# Nomination-Only Feedback Rescue

2026-10-04. Frozen AFTER consumed FITfeedback-v1 negative, BEFORE this rescue's
predictions. SameALL5183documents/351FITqueries/221units/5frozen held-fold models.
No NF/CAL/CONFtuning or predictions. Hypothesis: query expansion can add useful
nominees, but changing the base score causes observed ordering drift. Alternatives:
original score cannot promote added targets, sourceIDF correction incompatible,
insufficient nomination quality. Falsifier: no quality gain or worse top10order
over BOTH originalBM25 and originalBM25+IDF in bothaggregations. No adoption.

Run original BM25cap200, SAMEfrozen feedback10docs/10terms/.5mixture/df<=.1,
expandedBM25cap200. Union IDs, original nominees first plus unseenexpanded IDs.
No label-based filtering. UNIONcap400. Recompute original-query scores for ALL
members using existing BM25 Rerank API in <=200batches; both batches score the
whole corpus, so repeated scans counted. Normalize once by global maximum,
Two additional whole-leg validation Rerank scans are charged to union time;
extract sourceIDF features for ALL<=400, Score existing Model API for EACH row,
sortALL, THEN select200 (and laterpack10). Not RRF, expandedscore weighting,
protected-top10/forced promotion or metadata authority. Original control nominees
remain eligible; no correction sees a pre-pruned fusion. Zero-model union is a
negative control: final200top10 must match original plain BM25 exactly.

The400candidate workload is explicit extra nomination work, not a200-frontier
cost claim. For each query original model score semantics/sourceepoch/weights
unchanged. Shared helper validates duplicate IDs, finite full-row scores/context,
empty/short/frontier caps; source/clock binding remains existing immutable input
contract. This is offline cached source scoring, NOT livehydration/native LSN.

Primaryscreen unionIDF must beat originalBM25 and originalBM25+IDF in recall10/
NDCG10 query+221equal-unit means, with no frontier recall loss; SAME previous
quality thresholds, all requested cases. All predictions/source/weights/model
orders independently audited BEFORE FITlabel evaluation. Race/vet/build/full
experiment and corruption controls, actual fullunion counters, source/corpus/
index/feedback repeated searches/features/Score/sort/traces/fsync/RSS costs.
Consumed development evidence only, needs DIFFERENT untouched agent tasks.
