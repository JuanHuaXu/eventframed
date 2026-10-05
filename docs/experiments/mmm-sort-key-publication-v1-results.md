# Sort-key publication v1: private phase-boundary gate passes

The frozen [protocol](mmm-sort-key-publication-v1-protocol.md) passes in an
isolated LibraVDB collection with a SQLite sidecar. This is a test-only Goal 6
component. It does not authorize production cutover or complete Goal 6.

## Evidence

- Opt-in command: `EVENTFRAME_RUN_SORT_KEY_PUBLICATION_V1=1 go test ./internal/store/libravdbstore -run '^TestResearchSortKeyPublicationV1$' -count=1 -v -timeout 3m`.
- The same test passes under `-race`; the package's ordinary tests and
  `go vet ./internal/store/libravdbstore` pass. Race-instrumented timing is
  excluded from the performance result.
- Protocol SHA256: `02d3df08c9501c9e67f86f1fd9f695b41de2c002eb332bde5e768dd2650b8fba`.
  Test SHA256: `4f2b36c6d46e1afe3dad929308ed27ac23626879d580a9d6112f260d1a3c8b77`.

| Phase or control | Result |
| --- | --- |
| PENDING before ALTER, after close/reopen | Denied |
| One of three rows backfilled, after close/reopen | Denied |
| All rows backfilled and scanned, before READY | Denied |
| READY bound to stable LSN, count, digest; after close/reopen | Accepted |
| Exact-LSN read at `.120Z` | Returned `.100Z` and `.120Z`; excluded `.125Z` |
| Absent sidecar | Denied |
| READY marker with wrong digest, current process and reopened process | Denied |
| Legacy unkeyed write after READY | Denied; full rescan refused publication |
| Marker changed to the post-write LSN with the old digest | Denied in the current and reopened processes |

The last control exposed a real prototype bug: keeping only the verified
digest in memory allowed a marker at a newer LSN to reuse an old certificate.
A failing regression reproduced the acceptance; the gate now stores and checks
the verified LSN as well as the digest. The repaired test and its race run pass.

The later [receipt audit](mmm-sortable-receipt-v1-results.md) also replaced
an invalid schema-presence check in the readiness helper. Publication and
reopen controls pass again with catalog-type and exact-LSN bind checks; the
test hash above names the unchanged publication test, not that helper.

The quiet READY gate made 1,000 sequential checks including two LibraVDB LSN
reads and one SQLite marker read. In the final ordinary run p50 was 6 us and
p99 was 31.75 us, below the frozen isolated 1 ms p99 ceiling. This is not
loaded Recall p99, writer throughput, or freshness. The race run reported
p99 172.542 us under instrumentation and is not used as a performance claim.

## Limits and next test

The test uses orderly close/reopen at declared phase boundaries, not an OS
crash or hardware power loss. It does not fence independent writer processes,
prove sidecar/LibraVDB atomicity, or maintain READY incrementally across
authorized sortable writes. Every new LibraVDB commit invalidates the marker;
that is safe but cannot sustain continuous learning. Next, freeze a private
authorized-writer and crash-phase protocol with receipt-bound marker renewal,
then measure full loaded service latency and feedback freshness. Production
remains untouched; Goal 6 and all seven whole goals remain open.
