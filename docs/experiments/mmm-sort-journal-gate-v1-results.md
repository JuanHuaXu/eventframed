# Journal-through-gate publication v1: private component pass

Date: 2026-10-02. Frozen [protocol](mmm-sort-journal-gate-v1-protocol.md).
This research-only method records a metadata-only LSN transition while
leaving the EventFrame count and hash-chain root unchanged. It requires one
exclusive gate owner and routing every Store write through that gate.

The functional test passed: new journal, exact duplicate, later EventFrame
batch, exact-LSN as-of read, reopen, conflicting duplicate, DB-only and
SQLite-complete interruption, and direct-journal-bypass denial. The direct
path remains independently [failed](mmm-sort-journal-coexistence-v1-results.md).
The race-instrumented functional test, ordinary store package tests and
`go vet` passed.

In one isolated run, the last 100 of 116 new-journal calls measured p50
5.019834 ms and p99 6.534708 ms. After those writes, 100 quiet
published-view reads measured p50 25.750 us and p99 47.709 us. These are
diagnostic costs, not offered-rate Service Recall or learner latency.

The gate infers the journal transition from bracketing LSN/snapshot reads
and exact journal readback. It has **no database-enforced writer fence or
exact journal commit receipt**. An interleaved raw metadata write without
snapshot motion cannot be distinguished by this method. The proof therefore
depends on an exclusive gate-only writer contract; a second independent
Store or out-of-band write is outside it. Cross-database crash atomicity,
large-corpus restart, full Recall/Search/Snapshot coherence, durable-label
freshness and production rollout remain untested. This is a Goal 6 backend
component, not Goal 6 completion. Production was not changed.

Reproduce:

```sh
EVENTFRAME_RUN_SORT_JOURNAL_GATE_V1=1 go test ./internal/store/libravdbstore -run '^TestResearchSortJournalGateV1$' -count=1 -v -timeout 5m
EVENTFRAME_RUN_SORT_JOURNAL_GATE_COST_V1=1 go test ./internal/store/libravdbstore -run '^TestResearchSortJournalGateCostV1$' -count=1 -v -timeout 5m
EVENTFRAME_RUN_SORT_JOURNAL_GATE_V1=1 go test -race ./internal/store/libravdbstore -run '^TestResearchSortJournalGateV1$' -count=1 -timeout 5m
go test ./internal/store/libravdbstore -count=1 -timeout 5m
go vet ./internal/store/libravdbstore
```

SHA-256: test file
`84acbb7655dd9fdf1ba8f759c9c303701054270509c3008af9fcc0cd0174ecbe`;
protocol `a32f74b2f1b3d44b5a36ef5357df5ee70bdaf02014dfdf647aea64068d403c47`.
