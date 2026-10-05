# Source-IDF Rank Features: Mean Screen Pass

2026-10-04. [Frozen protocol](scifact-idf-pairrank-v1-protocol.md). Complete
FIT experiment and independent audit passed. The weighted source features
improve both declared controls under both aggregations, but the recall gain is
still sparse and FIT is already consumed. No untouched or whole-goal completion
claim; all seven goals remain OPEN.

## Actual Comparison

All5183verified source records, all351FIT queries/221families/full200frontier,
same five family folds, labels, .25tanh eight-weight residual and optimizer.
Only title/body/bigram/numeric overlap features acquire corpus-IDF weights;
length features and masked native channels stay unchanged. DF counts the
title/body union once per document. Absent query terms remain in denominator;
future sources/unknown epoch fail before any statistics are used. This is
rank correction, not calibrated confidence or a forecast-law change.

| Query-Macro Arm | Recall10 | MRR10 | NDCG10 |
| --- | ---: | ---: | ---: |
| Common-frontier BM25 control | 79.81956% | 0.631353 | 0.667185 |
| Frozen unweighted source-only correction | 80.00950% | 0.635349 | 0.670806 |
| IDF source correction | 80.10446% | 0.638209 | 0.673732 |

IDF gain over BM25: +0.28490 recall percentage points, +0.006547 NDCG. Gain
over unweighted: +0.09497 recall points, +0.002926 NDCG. Family recall control
79.78381%, unweighted79.85923%, IDF80.01006%; family NDCG0.681731/0.685992/
0.688588. Common frontier stays92.92498%recall; pureBM25frontier92.64008%.
Every baseline full ranking is bit-identical to the frozen unweighted study.

Against BM25, only one family improves recall and220tie, with no recall loss.
The recovered FIT query is ID4, now packed position10; 20families improve
NDCG,7worsen,194tie. Against unweighted, recall improves onefamily/220ties;
NDCG15better/8worse/198ties. The mean screen PASS is not a confidence bound,
population generalization or a reason to relax the failed prior calibration
gate. Public citation relevance is not truth or agent usefulness.

## Correctness and Cost

`research/public-task-pilot/scifact-idf-pairrank-v1/audit-results.json` is
authoritative.18-source complete local test closure, all5core commands terminal
codezero (race, vet, build, experiment, benchmark), all11race roots execute
three times. Independent Node audit reconstructs all35936corpus df terms,
70200vectors, six independent model fits, all351complete held-fold rankings,
control origins and five corruption rejections. Inputs/targets byte-identical
to original all-pair study; every held family excluded from its fitted model.
No additional CAL/confirmation predictions; previously consumedCAL180 remains
consumed globally. Original negative learners and prior calibration uncertainty
are preserved, not overwritten by this pass.

Source-index build259.858ms, versus97.740ms unweighted. Feature extraction for
200candidates median0.913ms/p951.518ms/p991.889ms; unweighted0.378/0.620/0.678ms.
Two-arm full ranking median33.50us/p9955.29us. ALL-FIT training496.818ms, with
the same75076pairs/24no-pair cases/200steps. Whole executable3.44s and maximum
RSS310.55MiB, versus unweighted3.08s/189.58MiB. The prototype duplicates raw and
weighted term structures; its memory/build cost is explicit, not free.
Standalone single-rank22-30us remains a component benchmark, not serving proof.
No native search timing, concurrent background publication, live invalidation,
delayed feedback or loaded100ms/freshness claim follows from this offline run.

## Next Requirement

Freeze the ALL-FIT model before a different, unused transfer cohort. NFCorpus
preparation uses all3633documents/all323official test queries, not consumed
SciFactCAL or a hand-picked source family. Its judgments are automatic link/tag
proxies and its owner permits academic use; keep data local and report those
limitations. The new protocol requires source-preserving conversion, all-source
DF/BM25/feature/order checks, graded and binary metrics, shared-source-component
aggregation and full measured prototype costs. Preparation is not an executed
transfer experiment or native-store integration.

Methodological sources: [Robertson & Zaragoza2009](https://www.staff.city.ac.uk/~sbrp622/papers/foundations_bm25_review.pdf),
[Joachims2002](https://www.cs.cornell.edu/people/tj/publications/joachims_02c.pdf).
Our normalized IDF feature combination is an isolated adaptation, not their
performance theorem or a canonical BM25F implementation.
