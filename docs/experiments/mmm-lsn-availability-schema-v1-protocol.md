# LSN availability-schema v1: frozen Goal 6 contract probe

The rolling-LSN v1 screen could not run because its SQL predicate could not
bind `available_at`. This test-only probe asks whether registering that
metadata field in a private LibraVDB tenant collection is sufficient. It
does not change the production collection schema, query path, or migration.

Create a private store and tenant collection with the same dimension,
cosine metric, HNSW and memory-mapping settings as normal, plus
`available_at: StringField`. Use the unchanged research batch writer to
insert one past-available and one future-available EventFrame. Capture an
exact durable old LSN. Insert one immediately available EventFrame and
capture a new LSN. Run the same parameterized vector SQL with
`WHERE available_at <= $available_by` at both LSNs, with an as-of time
between the past and future availability. Expected IDs: old={past},
new={past, visible}; future must never appear. Compare with current
`Store.Search`, close and reopen the database, and repeat the new-LSN SQL.
Require both LSN pins to close with zero active leases. No query rewrite,
Go-side postfilter, or discarded failing case is allowed inside this probe.

Pass means only that a declared metadata schema can express the necessary
availability contract in this release. It does not establish safe migration
for existing collections, concurrency, latency, or Goal 6 freshness.
