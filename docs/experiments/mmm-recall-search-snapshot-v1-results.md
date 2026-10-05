# Recall Search/Snapshot interleaving v1: failed control

Date: 2026-10-02. Frozen [protocol](mmm-recall-search-snapshot-v1-protocol.md).
The real `Service.Recall` searched one as-of-visible event, then an opt-in
wrapper ingested another as-of-visible event before Recall read `Snapshot`.
The wrapper injected only once, so a detected stale read could retry.

Both memory and LibraVDB stores **failed** the consistency condition.
Each returned after one search attempt with only `seed` in the packet,
while the packet and durable Bayesian journal both claimed runtime version
3, which includes `interleaved`. Before the injected write the version was
2. This is a silent searched-candidates/snapshot mismatch, not a demonstrated
future-data leak. `recall_k=pack_k=50`, so a small retrieval budget did not
cause the omission.

The earlier `TestResearchSnapshotInterleavings` still passes; it places
the write between snapshot assembly and journal commit. That sibling guard
cannot detect an as-of-visible write placed **before** the snapshot is
captured. The test-only [pre-search pin](mmm-recall-presearch-snapshot-v1-results.md)
checks the ordering hypothesis separately. Preserve this failing control.

Reproduce the expected failure:

```sh
EVENTFRAME_RUN_RECALL_SEARCH_SNAPSHOT_V1=1 go test ./internal/service -run '^TestResearchRecallSearchSnapshotV1$' -count=1 -v -timeout 5m
```

SHA-256: shared diagnostic test
`7fd05da162430b780b1d24f658c21e0627908ac447135e2dfa437acdbd036e69`;
protocol `b5f60a65a8355e0093a39d8ce6c9c912b04558cfa8f6c5d42619f416fce2d2cf`.
