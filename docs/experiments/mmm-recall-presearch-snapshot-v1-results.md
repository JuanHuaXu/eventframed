# Pre-search snapshot pin v1: finite ordering mechanism pass

Date: 2026-10-02. Frozen [protocol](mmm-recall-presearch-snapshot-v1-protocol.md).
This is a test-only `EventStore` wrapper; production `Service.Recall` is
unchanged. It presents a snapshot captured just before each Search to that
attempt's subsequent snapshot read. The same as-of-visible write is injected
after the first Search. The existing journal guard rejects the first attempt;
Recall retries and searches the now-visible second event.

Both memory and LibraVDB arms passed. Each made exactly two search attempts,
returned `seed` and `interleaved`, and persisted a frontier journal whose
runtime version 3 matched the packet and current Store snapshot. The focused
race run including `TestResearchSnapshotInterleavings`, ordinary service
package tests, and `go vet` passed.

This proves only a deterministic two-record ordering mechanism. Capturing a
snapshot before Search is not a complete immutable read view: external
candidate retrieval, graph/certificate/posterior reads, old published-LSN
search results, and concurrent writes still need a shared authority or
explicit compatibility check. The journal-through-gate metadata transition
also remains test-only. No loaded latency, durable-label freshness, large
corpus, or production serving conclusion follows. Goal 6 and all seven
whole goals remain open; production was not changed.

Reproduce:

```sh
EVENTFRAME_RUN_RECALL_PRESEARCH_SNAPSHOT_V1=1 go test ./internal/service -run '^TestResearchRecallPreSearchSnapshotV1$' -count=1 -v -timeout 5m
EVENTFRAME_RUN_RECALL_PRESEARCH_SNAPSHOT_V1=1 go test -race ./internal/service -run '^(TestResearchRecallPreSearchSnapshotV1|TestResearchSnapshotInterleavings)$' -count=1 -timeout 5m
go test ./internal/service -count=1 -timeout 5m
go vet ./internal/service
```

SHA-256: shared diagnostic test
`7fd05da162430b780b1d24f658c21e0627908ac447135e2dfa437acdbd036e69`;
protocol `9a9d9b69c180037d99696a22c58f5bc420cd232c2599b850a2e018fd98d9c45a`.
