# Published-LSN Service Recall v8: deterministic event-only contract passes

Date: 2026-10-02. The [frozen protocol](mmm-published-recall-v8-protocol.md)
passed in a private 256D service fixture. A test-only `EventStore` adapter
bound one immutable certified `(LSN, runtime snapshot, published_at)`
view to each `Service.Recall` call. Full EventFrames and cosine scores came
from that LSN; authorized Bayesian frontier journals advanced through the
metadata-only publication gate. Production/default runtime was untouched.

| Interleave after first Search | Search attempts | Stale rejections | Result |
| --- | ---: | ---: | --- |
| One as-of-visible event | 2 | 1 | Retry nominated all three eligible events; packet and durable journal used runtime version 5; gate READY |
| One future-only event | 1 additional | 0 additional | Old version-5 pin completed without future leak while current runtime became version 6; packet and journal agreed; gate READY |
| Later Recall on new pointer | 1 additional | 0 additional | Still excluded the future event at the same as-of; packet and journal used the new snapshot |

The first attempted run failed **setup**, before the interleaving controls:
`Service.New` legitimately bound its Bayesian policy and moved LibraVDB
LSN 21 to 24 after the gate had been declared READY. The gate correctly
flagged that as an out-of-band write. The corrected fixture initializes
the service and binds its policy while the marker is PENDING, then
publishes and initializes the journal once. The frozen interleaving
vectors, hooks, caps, and success gates were not changed. This startup
ordering is required by the single-owner contract; it is not evidence
that an already-READY gate tolerates raw policy binding.

The adapter checked `JournalSnapshotCompatible` under its single-owner
lock before entering the journal gate. Thus the expected visible-write
stale rejection did not poison an otherwise current gate. A future-only
write remained compatible, as the existing journal rule allows. Focused
normal/race tests, ordinary package tests, and vet passed.

This is not a loaded multiworker or complete state-isolation result.
Graph, posterior, residual, certificate and policy states were held
constant after initialization. The research adapter is not wired into
production and does not prove external writer ownership, all-read
snapshot semantics, label-to-forecast freshness, agent usefulness, or
sub-100-ms full-service p99. Goal 6 and all seven whole goals remain open.

Reproduce:

```sh
EVENTFRAME_RUN_PUBLISHED_RECALL_V8=1 go test ./internal/store/libravdbstore -run '^TestResearchPublishedRecallV8$' -count=1 -v -timeout 5m
EVENTFRAME_RUN_PUBLISHED_RECALL_V8=1 go test -race ./internal/store/libravdbstore -run '^TestResearchPublishedRecallV8$' -count=1 -timeout 5m
go test ./internal/store/libravdbstore -count=1 -timeout 5m
go vet ./internal/store/libravdbstore
```

SHA-256: test `f417e899273dff776fd43329bf66986bcc2331df516def2ac6be546502be0447`;
protocol `782951b15f4e20a47bc5d3154ba3d9572fc6c1dbfa1e9787db9706d586fa4128`.
