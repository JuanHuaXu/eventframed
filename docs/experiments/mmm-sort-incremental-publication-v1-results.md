# Incremental sort-key publication v1: single-owner component passes

**Subsequent audit:** A same-Store write interleaved between the precheck
and receipt could be silently included under this version's marker. A
deterministic regression failed, and the final test-only single and batch
gates now compare paired LSN/snapshot motion; see the
[batch publication follow-up](mmm-sort-batch-publication-v1-results.md).
The isolated timings below are historical pre-guard measurements, not
final-code timings.

The frozen [private protocol](mmm-sort-incremental-publication-v1-protocol.md)
passes its finite functional and isolated-cost screens. This is a
research-only Goal 6 component behind a **single long-lived Store owner**,
not a serving cutover. The earlier [dual-open probe](mmm-sort-writer-ownership-v1-results.md)
shows why that ownership assumption matters.

## Contract and controls

The test-only gate begins with a full-scan three-row genesis certificate.
It stores an ordered SQLite journal of hashes over each durable EventFrame
payload, exact UTC sort key and vector. The READY marker carries the journal
chain root, row count and exact LibraVDB LSN. An authorized append holds the
local gate lock, checks the old marker against LibraVDB, commits via the
receipt-returning sortable writer, validates only the new row, then commits
the journal entry and new READY marker in one SQLite transaction. Capture
still reads two LibraVDB LSNs around one SQLite marker read. Reopen performs
a full journal-to-database verification and row-count check.

| Control | Result |
| --- | --- |
| Two new rows and exact duplicate | READY advances twice; duplicate adds no commit or marker change |
| Subsecond as-of before and after reopen | Three past rows returned; future `.125Z` and `.130Z` excluded |
| Stop after LibraVDB commit, before SQLite commit | Current and reopened gates deny |
| Stop after SQLite commit, before in-memory install | Current gate denies; reopened gate verifies and accepts |
| Corrupt one journal row hash | Reopen denies |
| Legacy unkeyed write through same Store | READY stales; later append refuses to launder it |
| Direct sortable write bypassing journal | READY stales; later append and reopen refuse silent adoption |

Normal and race-instrumented functional runs pass with no race report. The
ordinary `libravdbstore` package suite and `go vet` pass. The opt-in commands
are:

```sh
EVENTFRAME_RUN_SORT_INCREMENTAL_V1=1 go test ./internal/store/libravdbstore -run '^TestResearchIncrementalSortPublicationV1$' -count=1 -v -timeout 5m
EVENTFRAME_RUN_SORT_INCREMENTAL_V1=1 go test -race ./internal/store/libravdbstore -run '^TestResearchIncrementalSortPublicationV1$' -count=1 -v -timeout 5m
EVENTFRAME_RUN_SORT_INCREMENTAL_COST_V1=1 go test ./internal/store/libravdbstore -run '^TestResearchIncrementalSortPublicationCostV1$' -count=1 -v -timeout 8m
```

## Isolated cost

After 16 warm-ups, 100 new one-event appends in the final run had p50
6.933 ms and p99 8.742 ms. One hundred quiet READY captures had p50
6.083 us and p99 134.333 us. Both pass the frozen isolated ceilings of
100 ms append p99 and
1 ms capture p99. These are one-run component timings, not confidence
intervals, loaded Recall p99 or end-to-end feedback freshness. Append work
depends on the new row and fixed-size journal update, not a corpus scan;
the reopen verifier remains O(N) in journal/collection size.
An earlier run reported append p99 8.043 ms and capture p99 10.708 us;
the larger final capture tail is retained rather than hidden.

Protocol SHA256: `c4bd64460279229a338d870b73a5da67b0fd3330733a1d6d9c3ddfbf684dbd25`.
Test SHA256: `8204e3c05292c524c6c0abeba3b0f5fb14399f89da4e53b049d352ff16480c53`.

This result does not authenticate arbitrary external writers or repair a
DB-only commit after interruption; it fails closed until separate recovery.
The chain is an integrity check under the declared single-owner model, not
a substitute for an adversarial authentication boundary. Next test startup
verification cost at larger row counts, design safe DB-only recovery, and
enforce exclusive Store ownership before loaded service trials. Production
is untouched; Goal 6 and all seven whole goals remain open.
