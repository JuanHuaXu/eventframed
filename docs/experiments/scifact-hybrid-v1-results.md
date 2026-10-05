# Hybrid Nomination: Strong FIT Lead, Not Whole-Goal Validation

2026-10-04. Authoritative independent evaluation:
`research/public-task-pilot/scifact-hybrid-v1/audit-results.json`. The10-source
complete local test-dependency freeze, five actual terminal code-zero commands,
five race roots x3,351 independent full-score/order reconstructions and10
corruption rejections passed. Original native artifacts remain hash-identical.
All seven WHOLE research goals remain OPEN/ACTIVE.

## What Was Actually Tested

ALL5183 source-pooled 5W1H documents and351 original FIT queries. No document/
query truncation, convenient missing-target exclusions or source-ID leakage
into tokenization. Frozen BM25 parameters and k=60 fusion were not tuned from
results. No label access by the predictor; separate evaluator opened only FIT
citation labels after complete predictions and terminal/structural checks.
Calibration180/confirmation300 stay unused. Native search is replayed from the
verified actual-store v3 trace; this is not a new native/agent serving run.

| Query-Macro Metric | Native Search | BM25 Within Native | BM25 Nomination | Equal-Weight Fusion |
| --- | ---: | ---: | ---: | ---: |
| Frontier recall |54.89%|54.89%|92.64%|92.92%|
| Recall10 |47.98%|53.47%|79.82%|66.62%|
| MRR10 |0.3852|0.4724|0.6314|0.4866|
| NDCG10 |0.4025|0.4815|0.6672|0.5236|

Within-native ranking retains the full nominated set, hence identical frontier
recall. Its top-ten improvement shows ordering headroom, but most improvement
comes from finding the missing citations in the first place. Fusion nominally
adds0.28percentage points frontier recall beyond BM25, while LOSING13.20points
recall10 versus BM25. Do not report fusion as universally better or discard its
negative ordering result. Its higher frontier recall may still help a separately
tested downstream ranker; that combination was NOT tested in this experiment.

For221 source-family-unit means, native/within-native/BM25/fusion recall10 is
48.88%/54.50%/79.78%/65.80%. Versus native, BM25 improves recall10 in84units,
worsens3, ties134; fusion improves51, worsens0, ties170. BM25 improves frontier
recall in97units and worsens1; fusion improves97 and worsens0. This finite FIT
diagnostic is not an untouched confidence interval or universal robustness result.
Family separation by cited sources does not establish topic independence.

Every fused frontier has200records, compared with median8native. Fusion evicts
mean10.94native records and adds163.29lexical-only records per query. The union
is at most400, explicitly capped to200; no claim that all union recall survives.
Predictor output retains the complete selected frontier before packing10.

## Cost And Implementation Audit

The immutable index has35936terms,685233postings and1265694tokens. Actual index
construction119.75ms,184.48MB cumulative allocation (NOT retained RAM). Full
offline predictor maximum RSS105512960bytes, about100.6MiB; native daemon memory
is separate. Each arm performs its own scoring, so all experiment costs are
counted, not attributed only to the chosen arm.

| New Offline Component | Median | p95 | p99 | Maximum |
| --- | ---: | ---: | ---: | ---: |
| Lexical nomination |0.705ms|0.826ms|1.142ms|1.341ms|
| Within-native BM25 |0.233ms|0.343ms|0.369ms|0.398ms|
| Fusion |0.030ms|0.059ms|0.065ms|0.075ms|

All351query phases including trace reading/output took611.46ms; full predictor
command986ms. Synthetic5183-document common-token search benchmark0.712-0.919ms,
427320-427323B/op and53allocations/op. This benchmark is a component stress
control, not billion-record scaling or complete serving. Query work scales with
posting frequencies and matched-document sorting; it is not constant time.

Replayed native search/binding plus measured lexical/fusion costs have a
counterfactual p95 of99.38ms and p99 of153.48ms. They were measured in DIFFERENT
processes/times: this is NOT observed live/loaded latency. It omits prospective
native hydration/binding of new lexical candidates, background contention,
publication freshness and real agent output. The sub100ms tail goal remains
unmet/unproven. Prior three native processes total607.997s; this extra offline
predictor does not erase that acquisition or earlier native failures.

Bug-hunt controls cover independently reconstructed formula/ties/Unicode,
full-frontier permutation, future exclusion BEFORE IDF statistics, immutable
input/output ownership, duplicate/unknown/nonfinite rejection, cancellation,
concurrent deterministic use, and output/source tampering. No future labels,
confidence rescaling or probability-law changes are introduced. Source body
text is exactly the stored source-pooled EventFrame text, not annotation metadata.

## What This Changes Next

Nomination and order both matter, and cheap lexical nomination is a credible
new lead. Next freeze a real pre-packing adapter with source binding/hydration
of lexical additions; compare a bounded learned correction against the stronger
BM25 control on a common200frontier. Only then dispatch previously untouched
outcome-labeled tasks under a frozen adoption rule. Test stationary harm,
shifts/delayed feedback and loaded freshness separately. A strong static
retrieval baseline does not establish EventFrame continuous learning, calibrated
beliefs, faster Anti-Pigeon cuts or any other whole research goal.

The implementation follows the declared positive-IDF BM25 variant and finite
RRF treatment, inspired by [Robertson and Zaragoza](https://www.staff.city.ac.uk/~sbrp622/papers/foundations_bm25_review.pdf),
[Cormack et al.](https://plg.uwaterloo.ca/~gvcormac/cormacksigir09-rrf.pdf) and
[BEIR](https://arxiv.org/abs/2104.08663). These sources motivate techniques,
not predictions that the present runtime will pass. Model task-prefix handling
and512-token consumption remain unproven native leads; neither was changed.
No production/private corpus/whitepaper edits, installs, commits, pushes or
deployments were performed.
