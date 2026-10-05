# Frozen NFCorpus Transfer: Small Mean Gain, Failed Gate

2026-10-04. All3633BEIR NFCorpus source documents/all323official test queries,
no NFCorpus fit or model selection. Frozen8-weight IDF all-pair model from
351previously consumed SciFact FIT cases. Protocol:
[pre-dispatch definition](nfcorpus-transfer-v1-protocol.md). No native service,
production changes or private corpus. This is a static public retrieval transfer,
not a complete outcome-labeled agent task or continuous-learning validation.

## Results

| Query Macro | BM25 Control | Frozen Correction |
|---|---:|---:|
| Frontier recall | 26.38548% | 26.38548% |
| Recall@10 | 14.77744% | 14.89072% |
| MRR@10 | 0.518781 | 0.518425 |
| Binary NDCG@10 | 0.306942 | 0.308987 |
| Linear-graded NDCG@10 | 0.307101 | 0.309312 |

Recall gain+0.11328percentage points; MRR slightly worsens. Grades are11758grade1
and576grade2 judgments. Use grade>0 for binary relevance; declared LINEAR grade
gain for graded NDCG, not an exponential-gain leaderboard comparison. Relevance
is extracted-link/tag proxy, not truth, clinical authority, causal support or
observed agent usefulness. Corpus/model pretraining exposure is unknown.

There are6positive-qrel overlap components, ONE with318queries, five singleton
components. No artificial query independence or splitting the giant component.
Equal-component recall41.26823% ->41.28741%; graded NDCG0.424828 ->0.425203.
Recall/NDCG improve only the giant component; MRR worsens only that component.
Frozen10000draw/20261005 whole-component bootstrap: recall difference95%
[0,+0.000575304], graded NDCG[0,+0.001123038]. Both lower bounds0 FAIL strict
positive promotion. These are coarse descriptive intervals, not proven
population/anytime coverage; six components, especially this giant component,
do not establish independent topics. NO PROMOTION; full goal5 still OPEN.

25queries legitimately nominate ZERO lexical matches; all retained with empty
rankings and zero retrieval metrics. Other frontiers may contain<200; all40470
nominated records are hydrated/scored BEFORE pack10. No query or source removed
to make quality, bounds or timings pass. Weak frontier recall is an upstream
nomination/lexical-gap limitation; bounded rank correction cannot retrieve an
absent source. Whether semantic expansion improves it at acceptable total cost
requires a separately frozen matched trial, not retrospective success credit.

## Audit and Cost

184-source complete local test dependency closure;4terminal-zero commands
(race,vet,build,experiment);26roots x3, no skips or data races. Legacy full-source
regression tests explicitly use their5183-document SciFact fixture, separately
frozen. Actual NF conversion/pooling and independent audit coverALL3633documents/
7612spans/5,782,778canonicalbytes, max2048byte field. Unknown who/when/why/how
stay unknown. Namespace research-scifact is unchanged auditor compatibility,
NOT a claim these are SciFact sources. Pool hash seals the actual epoch;
corpus DF excludes outcome labels and includes only as-of available documents.

Independent audit reconstructsALL40470feature vectors/323BM25nominations/two
complete rankings, all source spans/pools and model/source/clock/order identities
BEFORE opening relevance values.36source/pool/ranking corruptions plus ONE
separate empty-frontier corruption rejected. Raw-source membership projection
earlier read query IDs from TSV, not used grades/doc IDs for training/nomination.
Separate reporting readback recomputes query/components/metrics/serial cost.

Conversion41.839ms, pool71.116ms (includes second conversion), BM25index88.091ms,
IDFindex189.810ms; setup570.828ms includes source outputs. Serial search/hydrate/
feature/two-rank stages median0.822ms/p951.604ms/p991.823ms/max1.970ms. Searchp99
0.524ms, hydrate0.722ms, features0.759ms, two ranks0.040ms. Query phase270.444ms
includes29.374ms measured query writes; whole-before-final-sync841.492ms, actual
process wall1.11s, maxRSS280477696bytes=267.484MiB. Frames33.759MB/pool10.709MB/
trace22.596MB. Not loaded serving, durable background freshness or a matched
speedup over the earlier different native cohort; many short/empty frontiers
make the workload materially different. Final fsync counted in process wall.

## Preserved Failures

Acquisitionv1 CRLF header mismatch repaired ONLY in newv2 source root, same
verified archive. Predictor trialv1 compile failure (byte decoder received a
file handle), v2 mechanical clone changed unchanged protocol path, v3 race
fixture env missing; all before predictions, no labels evaluated. Saved source
copies/failure records. Trialv4 is actual terminal predictor; original frozen
auditor failed a vacuous empty-array pop control before relevance read.
NEW audit-v2 chooses a nonempty corruption case and checks empty frontiers;
same immutable predictions/model/protocol, no rerun or tuning. Generation receipts
and new auditor hash supplement, not rewrite, original frozen source closure.

Evidence: `research/public-task-pilot/nfcorpus-transfer-v4/{manifest.json,
commands.json,predictions.ndjson,audit-results.json,readback-results.json}`;
data lineage in `nfcorpus-v2/source.json`; prior failed trial roots retained.
Keep local academic use only; no redistribution/commercial license inference.
Primary dataset/method source: [Boteva et al.2016](https://www.cl.uni-heidelberg.de/~riezler/publications/papers/ECIR2016.pdf),
[official BEIR table](https://github.com/beir-cellar/beir).
