# Pre-search snapshot pin v1: frozen test-only rescue

Use the same two-store, two-record interleaving as the failed
[Search/Snapshot diagnostic](mmm-recall-search-snapshot-v1-protocol.md).
A test wrapper captures the underlying Store snapshot immediately before
`Search`, then returns that captured snapshot to the one subsequent
`Service.Recall` snapshot read. After the first Search has completed, inject
one as-of-visible event as before. Do not modify production service code.

The existing journal compatibility guard should reject the first attempt
because the injected event is visible at `as_of`. Recall should retry once,
search both records, publish a packet and journal with the current snapshot,
and include both event IDs. Run against memory and LibraVDB stores; verify
the already-existing Snapshot-to-Journal interleaving test still passes.

This is an ordering mechanism probe, not a coherent published-LSN adapter.
It does not certify external nomination, graph/certificate/posterior reads,
concurrent load, or production rollout. Preserve the failed original test.
