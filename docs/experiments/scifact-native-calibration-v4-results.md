# Fresh Native Calibration Retrieval

2026-10-04. Actual complete trial and independent audit passed correctness.
The frozen primary improves mean retrieval on new calibration queries, but
FAILS the predeclared promotion screen because recall uncertainty reaches zero.
Confirmation remains closed. All seven whole goals remain OPEN.

## What Ran

[Pre-dispatch protocol](scifact-native-calibration-v4-protocol.md). All 5183
documents, all 180 previously unforecasted calibration queries, 104 source
families. The primary all-pair and secondary Lambda source-only models were
frozen from exactly the original 351 FIT queries. No refitting, model selection
on calibration results, labels in predictor inputs, native-ranker credit, corpus
shrinking or convenient query exclusions.

A new owned COPY of the closed full-import LibraVDB store was launched with the
same installed Q8 GGUF model, explicit Unix socket, priority 10 and sampled 2GiB
stop ceiling. Typed ListByMeta reread and source-bound every one of the 5183
stored records before any new prediction. A SHA256 of that exact raw verification
prefix sealed the source snapshot epoch. Native SearchK200 plus BM25K200 nominates
a 60-RRF frontier, capped at 200; all 200 stored 5W1H payloads are hydrated and
verified before BM25 normalization and bounded source-feature rank correction.
Packing to ten happens only in the evaluator.

These are actual stored source frames, not assumed import acknowledgements or a
production text-corpus proxy. The cache is valid for this owned read-only trial,
not certified against concurrent mutation. Ranking scores are not calibrated
confidence or a changed probabilistic forecast law.

## Fresh Query Results

| Arm | Frontier Recall | Recall10 | MRR10 | NDCG10 |
| --- | ---: | ---: | ---: | ---: |
| Actual native search | 53.33333% | 45.55556% | 0.402515 | 0.409376 |
| Pure BM25 | 91.66667% | 77.77778% | 0.639195 | 0.670023 |
| Common-frontier BM25 control | 91.66667% | 77.77778% | 0.639195 | 0.670023 |
| Frozen primary source-only correction | 91.66667% | 78.33333% | 0.648269 | 0.678251 |
| Frozen secondary source-only correction | 91.66667% | 78.33333% | 0.647066 | 0.677280 |

Primary query-macro recall improves 0.55556 percentage points and NDCG improves
0.008227. Equal-family recall improves 80.19231% -> 81.15385%; family NDCG
0.693857 -> 0.704601. Source-family means, not individual queries, determine the
paired bootstrap. No top-ten target was lost by either learned arm here, but
ranking quality worsened in two families.

The recall gain is one recovered query (ID 1247), in one source family; 103
families tie and none worsen recall. Its relevant source was already inside the
frontier, and correction moves it into packed position ten. Eight families
improve primary MRR/NDCG, two worsen, 94 tie. This is a legitimate fresh ranking
gain, but not evidence of widespread improvement or answer correctness.

Primary family-paired differences with fixed-seed 10000-draw 95% percentile
bootstrap intervals:

| Metric | Mean Difference | Interval |
| --- | ---: | ---: |
| Recall10 | +0.009615 | [0, +0.028846] |
| MRR10 | +0.011258 | [+0.001242, +0.026723] |
| NDCG10 | +0.010744 | [+0.001191, +0.023539] |

Promotion required positive recall10/NDCG means in both aggregations AND a
strictly positive family-bootstrap lower bound for both. Recall's lower bound
is zero, so the decision is NO PROMOTION, not validated recall superiority.
The intervals are descriptive under exchangeable-independent family assumptions;
source separation does not prove topic independence. They are not population
certificates or anytime confidence sequences. Secondary is reported, not chosen
as a post-hoc replacement primary; it has the same recall limitation.

## Audit and Outcome Boundary

Authoritative `research/public-task-pilot/scifact-native-calibration-v4/`
contains frozen source closure, models, model origins, calibration-only queries,
command records, logs, actual binary, native-manifest link and audit results.
Actual raw trace and owned stores live in
`research/public-task-pilot/nativecal-v4/`. Original stores were rehashed after
the copy trial and stayed unchanged.

All four core commands (race, vet, build, native) exited zero. Complete local
test-dependent source closure: 45 files. All 19 specified race roots actually
executed three times without skips or data-race warnings. The independent Node
auditor validates all stored rows, exact prefix epoch, two FIT-only model origins,
all 36000 feature vectors, all 180 full-frontier nominations/hydrated payloads
and three complete rankings per query. Ten corruptions reject: dropped frontier,
reversed control, changed score, wrong epoch, wrong partition ID, altered text,
future metadata, duplicate identity, masked-channel reactivation and changed
lexical score.

Only AFTER these checks and client/daemon termination did the evaluator open
calibration citation targets. All 180 now count as consumed calibration outcomes;
they cannot be reused as untouched confirmation. All 300 confirmation forecasts
remain unused and official test labels were not opened. Public-pretraining
exposure remains unproven; citation relevance does not label support, truth,
contradiction, usefulness or real-agent delayed feedback.

## Actual Cost and Performance Limit

Serial serving-stage sum: median 93.710 ms, p95 103.209 ms, p99 173.816 ms,
maximum 257.764 ms. Native search alone: median 90.913 ms, p99 170.302 ms.
The additional mean stage cost is about 2.984 ms. Individual p99s cannot be
added to estimate a request p99; the stage-sum distribution is computed per
actual query. No loaded-serving or sub-100ms tail success follows.

Lexical nomination/fusion/rerank p99 1.844 ms; hydration and source binding
1.533 ms; features/masking 0.840 ms; three full rank arms 0.130 ms. These costs
include the validation work, not just an isolated score dot product. They exclude
trace serialization/fsync. Whole predictor time includes those journal costs,
verification and indexing: 72.377 s. The owned daemon wall time is 73.971 s;
500ms sampling saw a peak 1067.19 MiB, not a hard maximum-RSS guarantee.

Setup: 5183 typed verification RPCs totaled 34.563 s; combined lexical/source
index construction 218.239 ms. Cached source text is 8.583 MB, actual native
metadata 11.998 MB, raw audit trace 288.282 MB. These byte counts are not heap
RSS or evidence of bounded live invalidation. Original full-FIT model training
costs 496.427/863.044 ms are separately retained; they are not retrained or
silently credited as free. Earlier failed native launches/imports and original
failed rank learners remain failed with their original costs preserved.

## Research Consequences

The source-only model now has a fresh mean-ranking benefit, rather than only an
in-FIT screen. The stronger BM25 control remains essential; comparison only to
weak native search would exaggerate the learned correction. More independent
outcome families are needed to establish recall benefit, not more repeats of
the one recovered family. Do not lower the gate or refit on these outcomes and
call the result confirmation.

Next quality lead: source-discriminative (for example IDF-weighted) features in
isolated FIT-only studies, then a different frozen untouched task cohort; preserve
these calibration losses and sparse recall uncertainty. Next performance lead:
a versioned source-bound lexical-first adapter, measuring when expensive semantic
nomination adds value and how its background work affects freshness/loaded
serving. A cached snapshot alone does not solve invalidation or publication.
Delayed evidence, recovery stability, useful split outcomes and equal-total-cost
observation remain separate whole-goal requirements, not completed by this trial.
