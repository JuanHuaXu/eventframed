# Published-LSN unit-vector cosine v3: frozen restricted rescue

The v2 cosine SQL query failed on a nonunit stored vector because the
database's SQL cosine operator assumes unit inputs while ordinary Search
normalizes vectors for its index. Test a restricted contract in a fresh
private declared-payload collection: every nonzero stored vector and the
query are unit-normalized before the SQL call. Include two aligned rows,
an angled row `(0.6,0.8,0,0)`, an orthogonal row, an opposite row, and a
future-available sentinel. At the certified LSN, require the SQL `<=>`
distance transformed by `clamp(1-distance,-1,1)` to match ordinary
`Store.Search` similarity for every visible ID within `1e-5`, with
coherent bodies and no future admission.

Separately test a unit-norm validator on the same query scaled to
`(2,0,0,0)` and on a stored `(3,4,0,0)` vector. The validator must
reject both before a SQL score can be trusted. This does not implement
normalization or migration in production and does not extend to arbitrary
external pre-embedded vectors. A pass is only a restricted backend score
component; a writer/read-boundary enforcement contract is still required.
