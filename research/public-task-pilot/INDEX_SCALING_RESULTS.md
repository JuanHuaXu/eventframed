# Transaction indexing explains a corpus-dependent write cost

The frozen diagnostic completed24 observations (two per arm), with source hashes
in index-scaling-results.json. No agent-started tests ran concurrently. Normal
libravdb v1.6.13, synchronous durability, Apple M4/CPU4. Timings below are means
of two runs, not confidence intervals. Setup and reopen are excluded.

| Existing events | Index |16 separate writes, ms | One16-write batch, ms |
| --- | --- | --- | --- |
|200|HNSW|144.413|10.037|
|800|HNSW|399.443|24.963|
|1600|HNSW|826.433|51.523|
|200|Flat|71.504|5.414|
|800|Flat|80.426|5.815|
|1600|Flat|79.757|5.857|

HNSW ordinary-write time rises about5.72x across an8x corpus increase. Its batch
time rises about5.13x. Flat ordinary-write time is approximately stable above800
in this small sample, and batch allocation stays about163KB versus HNSW's
1.13MB->2.36MB. The source inspection's rebuild path is consistent with these
measurements. This does not establish a universal asymptotic fit or disk ceiling.

Flat trades write cost for search work. At1600 events, measured mean self-query
search takes approximately0.775-0.784ms for flat versus0.348-0.350ms for HNSW,
depending on write arm. All384 self-vector queries found their own ID in top10.
These are pseudo-random32-dimensional vectors, not semantic embeddings; this is
not evidence of general recall parity, especially at millions of events.

Every run checks snapshot increments and reopens to verify the final snapshot
and all16 inserted payloads. These are correctness guards, not crash tests.

## Decision

Keep investigating transactional indexing, not just queue scheduling. A small-
collection flat configuration is a plausible diagnostic or explicitly bounded
deployment choice, not a replacement for scalable ANN retrieval. Before changing
the backend, inspect current upstream fixes and options for transactional
incremental HNSW or snapshot-index publication. Such a design must preserve
search visibility, durability and snapshot validity on failure.

The separate queue improvement remains useful: a provably never-entered admission
failure should not poison later writes, while an uncertain commit must. Neither
that change nor these isolated timings establish a passing high-rate service.
No production/module/whitepaper changes. All seven whole goals remain open.

Verification: `node research/public-task-pilot/check-index-scaling.mjs`.
