# Immutable base plus bounded delta: research candidate

The direct HNSW incremental adapter is not ready to implement: the pinned
PreparedMutation contract explicitly requires Commit to publish materialized
state without allocating or searching. Calling HNSW.Insert there would move
fallible graph construction after durable commit and violate that contract.
Preparing by mutating the live graph would instead expose uncommitted changes.
Neither shortcut is a valid rescue.

Candidate: retain an immutable ANN base and prepare a bounded immutable delta
before the WAL commit. After durability, publish a preallocated generation
pointer. Queries search the base and score the bounded delta, shadowing every
changed/deleted ID. Async compaction builds a new base, but publication must
include or preserve all writes newer than its captured generation. Old readers
retain the old generation until safe reclamation. At capacity, use explicit
backpressure rather than unbounded growth or hidden stale search.

## First executable invariant

internal/researchindex/merge.go implements only the merge contract, not storage,
ANN construction, publication or compaction. With query-consistent comparable
scores, k requested results and d delta IDs, obtain at least the best k+d base
candidates (or the complete base if smaller), remove all shadowed IDs, add live
delta candidates, then select top k using a declared ID tie break.

For an exact base ranking, at most d prefix entries can be shadowed, so any
unshadowed base result outside that prefix cannot enter top k. The same operation
on an approximate ANN prefix does not create an exact-recall guarantee. Missing
prefix candidates cannot be recovered or certified by the merge function.

The test compares2000 generated insert/update/delete/tie cases against a full
exact oracle. A negative control demonstrates that requesting only k base
candidates can lose the answer after tombstoning or demoting leading entries.
Limits are k<=200 and d<=128 in this prototype, not a production tuning claim.
All tests pass three race-enabled repetitions (`go test -race
./internal/researchindex -count=3 -timeout=60s`). Repetitions reuse the same
generated cases; they are not additional independent semantic evidence.

## Remaining work before a backend candidate is usable

Follow-up: [generation lifecycle reference](GENERATION_PUBLICATION_RESULTS.md)
now exercises ownership, publication and compaction in memory. The following
integration requirements still apply to the real backend, not just that model.

Generation ownership, prepare/abort semantics, no-failure publication, CAS or
writer exclusion across durable commit, safe memory reclamation, recovery,
concurrent compaction with intervening updates/deletes, bounded compaction
backlog, inherited ANN quality, public task parity and full service latency all
remain unimplemented here. Search cost includes ANN overfetch plus bounded
delta scoring; compaction still has corpus-scale cost and cannot be wished away.
Do not count this merge component as a completed goal or a throughput rescue.

## Latest write reconciliation

phase-reconcile-results.json checks copies of all32 phase-aware stores: all74
failed writes absent, all438 acknowledgements present,74 retries successful,
all seeds/search memberships/snapshots consistent, source hashes unchanged.
This includes12 index-stage timeout members. It proves only orderly recovery
for those runs. Validate with check-batch-reconcile.mjs using the phase recovery
and phase-queue-results.json paths as its two arguments.

All seven whole research directions remain open. Production remains unchanged.
