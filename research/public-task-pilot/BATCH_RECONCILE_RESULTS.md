# Batch timeout reconciliation and transaction-scaling lead

## Observed recovery

All32 source stores were copied before opening; their SHA256 values remain
unchanged. Source artifacts: batch-queue-clean-results.json. Recovery artifact:
batch-reconcile-results.json. Existing reconciliation runner used unchanged.

- All91 failed writer IDs were absent after orderly reopen, including all five
  members reporting index-insertion timeout errors.
- All421 acknowledged writer IDs were present with matching payloads.
- All6400 seed IDs were present; direct lookup and full-corpus search agreed.
- Before retries, each snapshot matched the insert-only fixture's event counts.
- All91 retries on copies inserted successfully, with no duplicate flag.
- Every copy ended with216 events, evidence epoch216 and runtime version217.
- All32 fresh recovery recalls succeeded.

`node research/public-task-pilot/check-batch-reconcile.mjs` validates source and
artifact hashes, identities, payload checks, counts, retries and snapshots.
This resolves these runs only, not all timeout outcomes or abrupt crashes.

## Source-level finding: whole-collection transactional HNSW work

The pinned module is github.com/xDarkicex/libravdb v1.6.13. The daemon's
store.collection uses WithHNSW(16,200,100), store.go near1715.

In the module's libravdb/tx.go:

- buildTransactionState near1806 checks index.DeltaIndex; non-delta collections
  call getAllVectors and populate base/working maps for the collection.
- prepareIndexDeltas near1452 uses incremental mutation only for DeltaIndex.
- buildIndexes near2114 walks the working collection, sorts entries, creates a
  new index and reinserts them.
- commitTxWithGraphReceipt near1184 builds these indexes before calling
  engine.CommitTx/CommitTxDurable near1340. Index errors return before that call.

In internal/index/interfaces.go the pinned implementation supplies
PrepareMutations on flatWrapper, not HNSW. In libravdb/collection.go near680,
buildIndexForEntries emits the observed "failed to insert vectors into index"
error after insertion into that new index fails.

Thus this transaction path has at least whole-collection traversal and sorting
for HNSW; neither its persistence cost nor its index work is bounded by the
retrieval frontier cap. This is source evidence for a scaling concern, not a
measured asymptotic curve or proof that it dominates every latency sample.
Batching amortizes rebuilds but does not eliminate their corpus dependence.
The pre-commit ordering explains why the five observed timeout members were
absent, but error-string matching is not a reliable runtime phase certificate.
Post-durable index-publication failures also exist later in this function.

## Next research actions

1. Distinguish a proven never-entered admission rejection from a commit error in
   the queue. Only the former can safely avoid the reconciliation stop without
   backend support. Test both paths and preserve the unknown-outcome guard.
2. Instrument or benchmark index work against collection size before choosing
   a remedy. Consider incremental transactional indexing or a declared small-
   collection flat index, with retrieval quality and scaling checks. A flat
   index passing this200-event fixture would not solve million-event retrieval.
3. Continue fixed-arrival testing with journal durability included. Do not
   reinterpret the passing isolated batch benchmark as whole-service success.

No production, dependency, paper or publication changes were made. All seven
whole research goals remain open.
