# Cross-tenant journal retry v1: frozen causal probe

Run ordinary `Service.Recall` for tenant A with one as-of-visible seed.
Immediately before its first `PutBayesianJournal`, ingest an event for
tenant B. Compare a B event visible at the request's `as_of` with a B
event available one second in the future. Run on memory and LibraVDB
stores. Both events must remain absent from A's candidate set.

The current global `JournalSnapshotCompatible` rule is expected to retry
for the visible B event and to accept the future B event without retry.
Record search attempts, packet/journal snapshots and candidate IDs. This
probe isolates global runtime motion as a source of cross-tenant retry;
it does not show that ignoring B is safe for every graph, policy, or
cross-tenant contract. No production change or latency claim follows.
