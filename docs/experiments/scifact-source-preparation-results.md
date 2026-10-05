# Scientific retrieval benchmark preparation

2026-10-03. Goal 5 preparation after the original metrology design gate proved
mathematically unattainable. This does not revise that failed experiment or
claim a new retrieval improvement.

The official [BEIR source](https://github.com/beir-cellar/beir) archive was
downloaded and checked against its published MD5. Local SHA256 is
`536e14446a0ba56ed1398ab1055f39fe852686ecad24a6306c80c490fa8e0165`.
It contains 5,183 scientific title/abstract records, 1,109 queries, 919 official
training relevance relations and 339 test relations. No private conversations,
package installation, production service change or publication was involved.

Although official query IDs are disjoint, 181 evidence documents occur in both
official partitions. Therefore all connected query/evidence-document components
containing a test query are reserved from fitting. All 300 official test queries
remain confirmation; their 278 connected training queries are excluded. Identity
hashing partitions the remaining components into 351 fitting queries (221 units)
and 180 calibration queries (104 units). Confirmation contains 247 components.
No evidence document crosses these three partitions. Topic independence is not
established by this graph check.

Prepared serving records exclude author metadata. Natural-language abstracts are
not thereby certified anonymous, and no redistribution rights are inferred from
the software license. Labels are separate evaluator/preparation files. They
measure document relevance, including possible contradictory evidence, not
truth, causal identification or Anti-Pigeon validity. Public pretrained models
may already know this benchmark: untouched means unused by our fitting/selection,
not unseen in model pretraining.

The producer uses union-find. The separate auditor reconstructs components by
breadth-first traversal, checks every partition and denominator, reverses input
order, and rejects twelve corrupted plan/relation variants. Artifact hashes bind
the audit to the actual source and split plan. See local `scifact-v1/source.json`,
`split-plan.json` and `split-audit.json` for exact inventories and hashes.

No embeddings, ranking comparison, outcome prediction or agent run has occurred.
The separate experiment cache now supports up to 12,000 role-separated entries
and 64 MiB of retained key/vector payload. This is not an RSS or runtime-overhead
bound. It refuses cold acquisition beyond either cap, returns owned vectors,
checks model identity on warm hits and after cold acquisition, and excludes
canceled/nonfinite results from cache publication. Three technical root tests
execute three times under race detection, plus vet; exact code, inherited runtime
source hashes and raw logs are preserved in `scifact-v1/cache-preflight`. These
are deterministic fake-embedder checks, not downloaded-data or latency results.
Before fitting: freeze the complete corpus, evidence-independent units, lawful
source-preserving post-contract frames, common nomination/retention/packing caps,
bounded embedding cache, total acquisition/import cost and independent metrics.
An improved ranker here would advance retrieval evidence, not finish all of Goal 5.
All seven whole goals remain open.
