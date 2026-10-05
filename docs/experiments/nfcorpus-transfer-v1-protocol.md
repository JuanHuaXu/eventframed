# Frozen Source-IDF Cross-Corpus Transfer

2026-10-04. Prepared before any NFCorpus model prediction or relevance evaluation.
All seven whole goals OPEN. Follow-up to the complete IDF FIT mean-screen pass;
SciFact calibration180 remains consumed and its recall promotion failed.
SciFact confirmation300 is not substituted for this new cohort.

Use the COMPLETE BEIR NFCorpus variant (3633documents, all323official test
queries), not the larger original full-text archive or a chosen subset. Primary
is exactly the IDF all-pair ALL-SciFact-FIT model (fold-1,351FIT queries). No
NFCorpus training/dev/test relevance fitting, threshold/feature modification
or model selection. Freeze corpus/partition membership/model/source hashes before
dispatch. Partition projection may inspect only official-test query membership,
not use relevance values/document identities to select cases or nominations.

Convert documents after the public source contract to source-preserving5W1H
assertion spans; unknown causal/agent fields remain unknown. Join spans into
one bounded document record using the existing pool contract. Exact source
byte/digest/span coverage; source-import clock is NOT publication time. Do not
drop lengthy records, fabricate fields or silently truncate them to pass bounds;
a bound failure stops the complete trial and is retained.

New standalone lexical-first prototype: full-corpus BM25top200, source payload
hydration before scoring, normalized base, frozen IDF model, complete frontier
preserved before pack10. Common200control is plain BM25; no semantic/native
candidate trace or native database claim. This tests a retrieval technique
transfer and measures actual prototype cost, not native-storage integration,
live invalidation, online learning, loaded background freshness or agent answers.

Predictor gets corpus, test-only query text, frozen weights and NEW outputs,
never a qrels path. Journal ALL323complete pre-outcome predictions. After source,
as-of, full-frontier, formula/model/identity/order/corruption checks and terminal
predictor, evaluator opens official test relevance. Binary retrieval metrics
use score>0 as relevant; report graded NDCG separately with linear relevance
gain equal to the declared grade (no silent claim all labels are binary).
Relevance is an automatically extracted link/tag proxy, not factual truth,
support/contradiction, clinical advice, or observed agent usefulness.

Form source-overlap components on the fixed positive-qrel bipartite graph after
predictions solely for evaluation aggregation. Report ALL query means and ALL
component means, including a giant component if present; no treating repeated
queries with shared relevant documents as independent families. Component
separation does not prove topic independence. Primary diagnostic screen requires
positive recall10 and gradedNDCG gains over common200BM25 in BOTH aggregations,
plus descriptive95%paired whole-component bootstrap lower bounds>0. Freeze
10000draws/seed20261005; if too few independent components, declare uncertainty
unsupported rather than substituting a query bootstrap. No positive result here
alone completes full goal5's feedback-authority/real-agent requirements.

Audit every pooled source and vector, all corpus dfs, BM25nomination, full323
rankings, absence of test targets from fitter/predictor, and deliberate epoch,
source-text, model, feature, missing/duplicate candidate corruptions. Measure
conversion/index/RSS/serialization/whole process and serial stage p95/p99,
not just a cached ranking microbenchmark. Publication/model-origin costs stay
separate and visible. Record pretraining/benchmark exposure as unknown.

Sources: [BEIR dataset table](https://github.com/beir-cellar/beir),
[NFCorpus owner description and terms](https://www.cl.uni-heidelberg.de/statnlpgroup/nfcorpus/),
[Boteva et al.2016](https://www.cl.uni-heidelberg.de/~riezler/publications/papers/ECIR2016.pdf).
The BEIR archive MD5 is a89dba18a62ef92f7d323ec890a0d38d. Owner grants academic
use; other NutritionFacts.org uses require contacting its owner. The HF card's
cc-by-sa-4.0 tag is not treated as overriding those owner terms. Keep this local
research preparation/data unpushed and do not imply commercial redistribution
authority or a legal conclusion about underlying medical abstracts.

Pre-dispatch implementation clarification: the frontier is at most200, not
artificially padded when fewer lexical matches exist. An empty nomination emits
two empty rankings and remains in the323-query denominator. Nonempty nominees
are ALL hydrated and scored. Use the existing research-scifact/public-import/
public-scientific namespace solely for unchanged source-auditor compatibility;
the corpus and partition identifiers remain explicitly NFCorpus. Snapshot epoch
is SHA256 of the exact pool.json bytes (JSON encoder including final newline).
Require exactly8 finite model weights, abs<=4, last two exactly0; no fitted
operation in the transfer predictor. The bootstrap screen is unsupported with
fewer than2 source-overlap components; component independence remains unproven
even with2 or more. All source/setup and trace write costs are reported.
