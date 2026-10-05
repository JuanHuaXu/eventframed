# Reconcile the clean batch-queue run

Run the existing research-durable-reconcile tool against
batch-queue-clean-results.json, writing NEW batch-reconcile-results.json and
private database copies. Do not modify the32 source databases. Verify source
SHA256 before/after. Use its original incremental service overlay: this omits
only the atomic residual allocator, whose output parity was separately tested.
This is a recovery audit, not a timed performance comparison.

For all512 attempted writer IDs: compare acknowledged/failed status to direct
post-reopen presence and original payload. Check all6400 seed records and compare
full-corpus search membership against direct lookup. Record snapshots before
retries and verify version/event-count agreement for this fixture. Execute a
fresh recall, then retry failed writes with their exact original CaptureTurn
IDs and payloads on copies only. Record duplicate flags, retry errors, final
presence and snapshot counters. Do not infer absence from a timed-out lookup.

No expected outcome is assumed for the five index-stage timeout members.
Present failed writes must retry idempotently; absent failed writes must insert
without duplication. Every acknowledged write must be present. This checks
orderly reopen recovery, not abrupt process failure or every partial-commit path.

All original artifacts remain intact. No production or whitepaper changes.
