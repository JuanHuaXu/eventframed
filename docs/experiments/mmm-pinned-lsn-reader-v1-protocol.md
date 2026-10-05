# Pinned-LSN reader v1: frozen Goal 6 design screen

This is a research-only reader-path experiment using the exact-LSN temporal
query API in the repository's pinned LibraVDB v1.6.13 dependency. It does
not modify the production `Store.Search`, publication, service, or daemon.
The question is whether a coherent pinned record/vector view avoids the
application-level event read lock while maintaining as-of validity.

In each fresh arm, seed 200 past-available same-tenant EventFrames, capture
the application snapshot and latest durable LibraVDB commit LSN, and pin
that LSN. Warm four reads outside timing. Offer 256 future-only same-tenant
events as 16 raw batches of 16 at nominal 16 ms batch intervals while eight
reader workers receive 192 queries at nominal 4 ms intervals. Compare the
current `Store.Search(k=200)` path with SQL vector search over the pinned
`AS OF LSN` relation. Rotate the two arms over three fresh matched pairs.

Every reader must return exactly the 200 seeded IDs and no future event.
Require `ResearchSnapshotCompatible(captured, asOf)` before and after every
read; the fixed `asOf` precedes all future arrivals. Confirm all 256 writes
and per-version motion. After the measured run, commit one event visible at
the fixed `asOf`: the compatibility guard must reject the captured snapshot
even though a pinned SQL query still returns the old 200. Close the pin and
verify no lease remains. Record read-call p50/p99, offer-to-response p99,
actual read offer gaps, batch writer completion and memory/lease stats where
available. Do not silently substitute a different SQL/vector query or omit
the guard to win a timing gate.

Frozen design screen: candidate pooled read-call p99 <=1.10x matched current
path, candidate pooled writer completion <=1.25x current path, zero guard
errors or future leakage, and all 256 writes/192 reads per arm. Report all
per-pair values. A pass would only qualify a fresh confirmation and a
service-level authority/integration study. Exact-LSN pinning may retain
history or rebuild a temporal index; report those costs. No result here
proves the full 4 ms Recall/learner freshness contract.
