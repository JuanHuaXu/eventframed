# Bounded leases and combined durable serving correctness

NewServingBounded now declares both retirement and active-lease caps; NewServing
uses a64-lease research default. Acquire rejects excess leases explicitly.
Release drops the historical view and graph references, so retaining a released
lease object does not pin those resources. Revision remains readable, and double
release cannot grant extra capacity. These limits bound counts, not total corpus
bytes or allocations held by arbitrary external code.

## Combined real-storage test

serving_durable_test.go uses an authoritative libravdb flat vector collection
plus a metadata revision record in synchronous transactions. The serving side
uses separate real immutable HNSW bases plus the bounded delta. This joins the
previously separate persistence and ANN correctness checks.

With a delta cap of4, the test inserts12 distinct-vector records, obtains the
expected top1 self-query after each acknowledged write, and repeatedly compacts
at capacity. It then atomically deletes one record and changes another. A retained
historical reader still answers from its original state. After orderly shutdown,
the authoritative store is reopened: the deletion, surviving records and revision13
are checked, and a reconstructed HNSW serving base returns the changed record.

The full researchindex suite passes three race-enabled repetitions:
`go test -race ./internal/researchindex -count=3 -timeout=120s`.
It includes the new lease-cap test and existing failure/publication regressions.
This controlled12-record test is not semantic recall validation or a throughput
experiment. Recovery still enumerates known fixture IDs, not arbitrary stores.

## Next requirements

Run a fixed-arrival workload with actual synchronous persistence, indexed reads,
compaction and all caller-visible errors/latencies. Compaction is currently
invoked explicitly on capacity and can build slowly; it needs a bounded proactive
scheduler to avoid capacity stalls. Avoid benchmarking only the cheap delta path
and excluding the cost required to keep it bounded over sustained operation.

EventFrame's real5w1h/event payload, idempotency, scored forecasts, durable journals
and public service replay still need integration. Flat authoritative indexing
is presently redundant with the derived HNSW; this fixture is not a production
storage-layout decision. No daemon, dependency, paper or publication changes.
All seven whole research goals remain open.
