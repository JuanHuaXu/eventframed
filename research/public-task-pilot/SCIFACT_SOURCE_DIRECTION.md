# External scientific retrieval source: preparation only

2026-10-03. New Goal5 lead after the consumed metrology DESIGN gate proves
headroom-impossible. No old cohort/model/threshold is changed, and no new
retrieval or agent result is claimed here.

Primary sources: [BEIR paper](https://arxiv.org/abs/2104.08663),
[official BEIR data/checksum table](https://github.com/beir-cellar/beir),
[SciFact project](https://github.com/allenai/scifact). Public benchmark acquisition
uses BEIR's explicit SciFact archive/checksum, not scraped private conversations.
No package installation or repository clone is necessary. The archive stays
local; code's license is not assumed to license all abstract text for arbitrary
redistribution. Separate text/data rights review is required before publication.

Corpus records contain public scientific titles/abstracts; author metadata is
excluded from the prepared serving corpus. This exclusion is NOT a blanket PII
or inference-anonymity certificate for natural-language abstracts. Do not publish
the downloaded data on that assumption. Query JSON contains only ID/text; qrels
are in evaluator-only files, not embeddings, rank inputs, EventFrames or caches.
IDs are identities, not predictive features. Acquisition validates identity,
query-split disjointness, binary qrels and corpus coverage before a model run.

These qrels label document RELEVANCE to a claim, including evidence that could
contradict it. They do NOT authenticate the claim as physical/medical truth,
identify a causal effect or certify Anti-Pigeon. Public pretrained embeddings
may have seen this material; future held-out claims refer to our fitting and
selection procedure, not a claim of absent model-pretraining contamination.

Official train/test query disjointness alone does not prove family independence:
queries may share evidence documents or topics. Measure that overlap before
declaring independent units. A future protocol must freeze grouped fitting,
calibration and untouched evaluation; retain all target misses, not just selected
baseline failures. Baseline/lexical/fusion/abstention arms must share EXACT
embeddings and nomination/frontier/packing bounds. Both real post-contract text
CaptureTurn and declared structured frames need source-preservation checks.
Do not smuggle a relevance label or gold rationale into frame extraction.

Keep law/rank separation, explicit cold acquisition and bounded embedding cache,
full journal/nomination/retention/packet audit, and independent metric/negative
controls. The former512-entry experiment memo cannot hold the full benchmark;
any larger experiment-only cache needs its own declared bound and cost, not a
silent change to runtime state or the frozen V44 sources.

[Conformal risk control](https://arxiv.org/abs/2208.02814) is a possible source
for a nested abstention risk gate, not automatic protection for arbitrary fusion
weights: ranking loss need not be monotone, and exchangeability/calibration
sample size must be justified. Earlier [RRF](https://cormack.uwaterloo.ca/cormack/cormacksigir09-rrf.pdf)
and lexical component wins/transfer failures remain evidence, not universal
fusion guarantees. Source preparation is separate from implementation/adoption.
All seven WHOLE research goals remain OPEN/ACTIVE; production stays untouched.
