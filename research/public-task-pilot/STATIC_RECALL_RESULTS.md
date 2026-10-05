# Static recall diagnostic results

Protocol: [STATIC_RECALL_PROTOCOL.md](STATIC_RECALL_PROTOCOL.md).
Artifact: `static-recall-results.json`; fresh derived stores retained beside it.
Runner: `cmd/research-static-recall`, frozen bulk-base overlay.

Both independent builds missed seed-954 in top 10, each over 1024 self queries
against 6400 records at 768 dimensions. There were no search errors or missing
owned reference vectors. Both returned seed-6128 first with cosine
0.13280827932487513, whereas the query's own vector has cosine 1.
All returned scores are independently recomputed by `check-static-recall.mjs`.
It also checks IDs, ordering, counts, hit flags and source hashes.

No mutations, delta entries, background compaction or concurrent requests were
present. Thus those mechanisms are not necessary for this reproduced recall
failure. The static adapter rescores and sorts the same ten nominated candidates;
without a delta it cannot drop the self candidate if ANN nominated it. The result
isolates a nomination limitation in this constructed base, not its internal cause
(construction, connectivity or search effort remain candidates).

This is not proof of the original growth run's exact failure transition. That
run did not retain candidate IDs or the graph generation for each read. In
particular, seed-779's original miss did not recur in these static builds.
The owned map check is not independent enumeration of the derived database.

Unloaded maximum query durations were 1.955083ms and 1.648ms. They do not replace
the growth run's open-loop latency results. This is diagnosis, not a passing
rescue or a semantic retrieval benchmark.

Next: compare search effort on the same static graph with a recorded candidate
trace and exact oracle; separately pursue tiered immutable runs to reduce whole-
corpus rebuild pressure. Do not hide either failure by relaxing the acceptance
criteria. All seven research goals remain open.
