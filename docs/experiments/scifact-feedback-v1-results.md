# Feedback Nomination and Original-Score Rescue: Both Screens Fail

2026-10-04. Development only: all5183framed source records/all351CONSUMED
SciFact FIT queries/221source families. Five already-frozen held-fold IDF models,
no refit, no NFtest/CAL/CONFpredictions or tuning. Two new isolated modules/commands,
original research and production files unchanged. Protocols:
[feedback](scifact-feedback-v1-protocol.md),
[nomination-only rescue](scifact-feedback-union-v1-protocol.md).

## Comparative Evidence

| Query Macro | BM25 | BM25+IDF | Feedback | Feedback+IDF | Union+Original+IDF |
|---|---:|---:|---:|---:|---:|
| Final200recall | 92.64008% | 92.64008% | 94.06458% | 94.06458% | 92.56885% |
| Recall@10 | 79.81956% | 80.10446% | 78.96486% | 79.98575% | 80.10446% |
| MRR@10 | .631353 | .638209 | .591926 | .604075 | .638209 |
| NDCG@10 | .667185 | .673732 | .638505 | .650332 | .673732 |

Bounded RM3-inspired feedback increases200nomination recall+1.42450percentage
points but damages answer ordering. Feedback+IDF MRR drops.034135/NDCG.023400
versus strongerBM25+IDF, with lower query-macro recall10. Equal-unit recall
80.01006% ->80.48140% improves, but MRR.657939->.611844 and NDCG.688588->.657788
worsen; no aggregation cherry-pick. ComponentMRR25better/47worse/149tie;
NDCG30better/46worse/145tie. Mean86.56of200original nominees displaced by the
EXPANDED leg. Uncertainty not needed to reject the frozen strict mean gate.
Feedback screen FAILS; no promotion or public task improvement claim.

Rescue separates nomination from scoring: union original200+expanded200, max400
allowed, actualmaximum366/mean286.56. All100584union candidates scored with
ORIGINAL-query BM25+sameIDF before200selection/10packing;170784feature vectors
including the351original controls. No original nominee dropped before scoring.
Raw union recall95.06173% is increased budget/headroom, NOT final200or10quality.
Four whole original-score passes per query (two leg validations plus<=2batches)
are charged, along with the feedback original-search recheck and expanded search.

Zero-model union's final200 ranking is EXACTLY the plainBM25 control for all351
queries, validating unchanged original scoring. Union+IDF top10IDs AND scores
are EXACTLY BM25+IDF for all351queries; same recall/MRR/NDCG, no rescued answers.
Final200frontier recall instead loses.071225percentage points in one family
(0better/1worse/220tie). Equal-family frontier93.44646% ->93.38989%. Mean3.658
original candidates evicted AFTER scoring/200selection. Rescue removes ordering
harm but adds no quality and mildly reduces final-frontier coverage; FAILS
unchanged gate. Primary expanded-score experiment and rescue negatives preserved.

The rescue audit's copied `nomination.meanEvicted` field describes the raw
EXPANDED leg versus original, NOT pre-scoring union loss. Independent readback
explicitly records0pre-score losses and the3.658final-IDFmean evictions. No
renaming/overwriting of frozen original audit files to hide that reporting caveat.

## Audit and Performance

Feedback23-source complete local test-dependent closure,18race rootsx3;
rescue27sources/20rootsx3. Each4terminal-zero core commands (race/vet/build/full
experiment), no skipped tests/data races. Independent audit verifies immutable
source-body/clock/epoch/title linkage to prior actual native-verified stored
records, all351whole nominators/feedback distributions,140400/170784feature
vectors and four complete rankings per query BEFORE consuming FITjudgments.
12deliberate corruption rejections each; scalar-model API reused, no new score
formula. Source/epoch/cancel/duplicate/invalid-term/mass/nonfinite/future/scored-
before-cap400bounds exercised by unit tests. Separate reporting check recomputes
all query/221family metrics, frozen screen, serial costs and exact top10parity.

| Actual Prototype Cost | Feedback | Union Rescue |
|---|---:|---:|
| Corpus-index build | 124.785ms | 123.042ms |
| Source-IDF index build | 259.196ms | 257.102ms |
| Serial stage median | 4.102ms | 5.498ms |
| Serial stage p99 | 7.145ms | 9.768ms |
| Serial stage maximum | 8.029ms | 11.603ms |
| Whole process wall | 2.18s | 2.66s |
| MaxRSS bytes | 366231552 | 359792640 |
| Output trace bytes | 79776579 | 97908064 |

Stage times include original nomination, feedback/recheck, expansion, union
validation/scoring where present, both feature frontiers/four rank arms. Query
writes102.646/130.226ms separately, finalsync in process wall. BodyTF retention
and redundant source maps/indexes counted. No native RPC/service, loaded serving,
online update/durability/freshness or200-only workload claim. Corpus is cached
source-bound5W1H text, not full-text agent memory; unknown fields stay unknown.
Citation labels are not truth or answer usefulness. Prior pretraining exposure
unknown; repeated consumedFITdevelopment is not fresh evidence.

## What This Changes

More nomination headroom ALONE does not make the current ranker use the new
sources. Expanded-score topic drift is measurable; preserving original scores
eliminates that drift but neutralizes the useful nomination gain at top10.
Need a stronger, independently validated bridge between nominated source and
query/outcome, or targeted source-bound semantic signals; neither existing
simple weighting nor bounded old residual delivers it here. Do not dispatch
untouched confirmation for these failed candidates or weaken original gates.
Live source/statistics invalidation/loaded freshness, delayed-score alignment,
useful split outcomes, robust challengers and equal-total-cost observation remain
viable separate leads. All seven WHOLE requirements OPEN/ACTIVE.

Evidence roots `research/public-task-pilot/scifact-feedback-v1` and
`scifact-feedback-union-v1`; independent `scifact-feedback-readback-results.json`.
Primary research inspiration: [Lavrenko & Croft2001](https://ciir.cs.umass.edu/pubfiles/ir-225.pdf)
and [official Anserini RM3](https://github.com/castorini/anserini/blob/master/src/main/java/io/anserini/rerank/lib/Rm3Reranker.java).
Unicode/unique-query/filter/model constants are this experiment, not a claim to
reproduce Lucene/Anserini or those papers' benchmark effectiveness.
