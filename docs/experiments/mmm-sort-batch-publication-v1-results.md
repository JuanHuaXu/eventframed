# Batch incremental publication v1: private component result

Date: 2026-10-02. Frozen [protocol](mmm-sort-batch-publication-v1-protocol.md).
This test-only single-owner component passes its finite functional controls
and isolated cost diagnostic. Production code and serving behavior are
unchanged. Goal 6 and all seven whole research goals remain open.

## Publication contract

One LibraVDB transaction commits up to 16 new EventFrames and returns one
exact commit LSN. The gate reads back each new row, adds one hash/sequence
entry per row to the SQLite journal, and atomically moves one READY marker
to the final count and chain root. Exact duplicates add no journal row;
duplicate-only retry neither commits nor advances the marker. Runtime version,
evidence epoch and ingestion motion still advance per new event. Every new
event in a batch becomes visible together at the batch commit LSN, so this
does **not** expose intermediate per-event snapshots or an immediate
per-event acknowledgement API.

All-new, exact-retry, mixed duplicate/new, conflict, exact-LSN subsecond
as-of, close/reopen verification, DB-before-SQLite interruption,
SQLite-before-in-memory interruption, sortable bypass and unkeyed bypass
controls pass. DB-only interruption denies READY after reopen; a fully
committed SQLite journal verifies after reopen. Normal and race-instrumented
functional runs, ordinary package tests and `go vet` pass.

## Interleaving audit

The first implementation passed sequential controls but **failed** a new
deterministic interleaving: an unkeyed write through the same Store after
the initial LSN check and before the candidate receipt was silently covered
by the later marker. The adjacent single-event gate had the same flaw.
Both failing regressions were observed before repair. The final test-only
gates now capture LSN and runtime snapshot together under the Store read
lock and require the post-commit snapshot to equal the pre-commit snapshot
plus exactly the journaled new events. Duplicate-only calls require no
snapshot or LSN motion. Both interleaving regressions then pass, normally
and under race instrumentation.

This protects same-Store writers that advance the Store snapshot. It does
not establish exclusive ownership across independent Store instances or
protect an out-of-band raw database mutation that bypasses the Store
snapshot. The earlier dual-open data-loss result still requires one
long-lived owner and a stronger operational fence before any cutover.

## Isolated cost

At each starting size, eight groups of 16 equal events were written to
separate private Stores. Control journaled each event separately;
candidate journaled the group after one batch receipt. Arm order rotated.
The table uses the final code with the interleaving guard.

| Starting rows | Single group p50 / p99 | Batch group p50 / p99 | Sum singles / batch | Batch-to-single total ratio | Quiet capture p99 |
| ---: | :--- | :--- | :--- | ---: | :--- |
| 259 | 187.448 / 209.993 ms | 12.858 / 16.928 ms | 1.542 s / 106.880 ms | 0.069 | 12.750 us |
| 1027 | 527.821 / 573.791 ms | 32.845 / 35.256 ms | 4.255 s / 262.237 ms | 0.062 | 13.708 us |

The batch substantially reduces *isolated persistence time per group*.
It does not prove 4 ms offered-rate freshness: group formation, waiting,
queueing, concurrent Recall, labels, full service calls and tail admission
were not measured. Batch acknowledgements occur only after the group
commits. The next meaningful Goal 6 experiment is a frozen open-loop
offered-rate test with one owner, exact event-level accounting and concurrent
as-of Recall; it must count dwell, backlog and p95/p99, not just DB call time.

Reproduce:

```sh
EVENTFRAME_RUN_SORT_INCREMENTAL_V1=1 EVENTFRAME_RUN_SORT_BATCH_PUBLICATION_V1=1 go test ./internal/store/libravdbstore -run '^(TestResearchIncrementalSortPublicationV1|TestResearchIncrementalSortBatchPublicationV1)$' -count=1 -v -timeout 5m
EVENTFRAME_RUN_SORT_INCREMENTAL_V1=1 EVENTFRAME_RUN_SORT_BATCH_PUBLICATION_V1=1 go test -race ./internal/store/libravdbstore -run '^(TestResearchIncrementalSortPublicationV1|TestResearchIncrementalSortBatchPublicationV1)$' -count=1 -timeout 5m
EVENTFRAME_RUN_SORT_BATCH_COST_V1=1 go test ./internal/store/libravdbstore -run '^TestResearchIncrementalSortBatchCostV1$' -count=1 -v -timeout 5m
go test ./internal/store/libravdbstore -count=1
go vet ./internal/store/libravdbstore
```

At-run SHA-256: batch test
`ba702743cc32e2adee5fbdefcc62b4bfc666bc5230303e5f6b1e4106748c658d`;
single test `ef4edbbbe600df2df444c751cfcb6a0208a2dcfdf7d1cceb031aa365033e7c37`;
protocol `84a43e7f206fd4afe1f2347b0374c0d3fc7128c33ec98d0d089f0345c05f1988`.
