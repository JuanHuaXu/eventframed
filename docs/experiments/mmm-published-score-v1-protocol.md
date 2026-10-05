# Published-LSN score semantics v1: frozen contrast probe

The payload projection test uses identical vectors and therefore does not
validate retrieval-score semantics. Create a separate private declared-
payload collection with three EventFrames at distinct vectors: one aligned
with the query, one orthogonal, and one future-available sentinel. Publish
and pin its exact LSN. Query the two as-of-visible rows with SQL
`embedding <-> $query_vec AS distance`, projecting the same payload fields.

Require exact EventFrame bodies, no future row, a finite SQL distance for
each result, and aligned-before-orthogonal ordering. Record the SQL
`SearchResult.Score`, projected distance, and ordinary `Store.Search`
similarity for each ID. The candidate passes as a direct Store.Search
replacement only if the SQL score agrees with ordinary Store similarity
within `1e-5`; otherwise preserve the mismatch and identify an explicit
calibrated conversion as further work. Do not infer score equivalence from
IDs or ordering alone. This is an isolated backend probe, not full Recall.
