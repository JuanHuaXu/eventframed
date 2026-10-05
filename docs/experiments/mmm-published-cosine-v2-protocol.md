# Published-LSN cosine conversion v2: frozen score rescue

The v1 direct-score contract failed because SQL `<->` is L2 distance and
SQL `SearchResult.Score` is lower-is-better distance, whereas ordinary
cosine collection Search exposes higher-is-better similarity. On a new
private declared-payload fixture, use the SQL cosine-distance operator
`embedding <=> $query_vec AS distance` at one certified published LSN.
Convert each projected distance to similarity as
`clamp(1 - distance, -1, 1)`; never treat SQL `row.Score` as similarity.

Use a unit query and five as-of-visible rows: two aligned controls, one
scaled angled vector, one orthogonal vector, and one opposite vector.
Include one future-available sentinel in the collection. Compare the SQL
IDs and converted similarities with ordinary `Store.Search` for every
visible row, tolerance `1e-5`. Require valid projected EventFrame bodies,
no future row, finite distances, and descending similarity order (ties
allowed). A mismatch fails the candidate. This is an isolated score
contract, not a full Service Recall or loaded publication test.
