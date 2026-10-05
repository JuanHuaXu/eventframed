# Static ANN miss diagnostic

This diagnostic isolates nomination in a newly built immutable base. It does
not reconstruct the original growth experiment's graph or repair its results.
No production service or retained growth database is opened.

- Two independent builds, 6400 SHA-derived 768-dimensional seed vectors each.
- Same HNSW parameters and frozen bulk-build overlay as growth.
- Sequential self queries for seed-0 through seed-1023, top 10, no mutations.
- Record every candidate ID and score, query duration, and exact owned-vector
  identity. A self-query has cosine 1; missing it is a recall failure, not proof
  of record deletion. Error counts remain separate from recall misses.
- No performance acceptance claim: this is unloaded diagnosis, not open-loop
  serving. Timing is diagnostic only.
- If misses occur here, concurrent publication and delta shadowing are not
  necessary causes. If none occur, the original causes remain unresolved.
- Preserve source hashes, all query results, and fresh derived databases.

The synthetic restored map is a reference fixture, not a durability test. The
persistence callback rejects all writes; the diagnostic never calls Apply.
