# Journal gate phase timing v14: LibraVDB write dominates

Date: 2026-10-02. The [frozen diagnostic](mmm-published-recall-journal-phases-v14-protocol.md)
timed an otherwise equivalent test-only copy of the journal gate on
v13's unchanged 4 ms, 256D, 128-write/128-Recall workload. A paired
test compared original and timed gate transitions for a normal
append, exact duplicate, and interruptions after the LibraVDB and
SQLite commits; all matched after reopen. Both normal loaded trials
completed 128/128 writes and 128/128 Recalls, with zero stale,
oracle, journal, future-leak or freshness violations. They still
**failed** the <100 ms offer-to-done p99 gate at 509.912 and
525.122 ms.

| Journal subphase | Normal trial 1 p50 / p99 | Normal trial 2 p50 / p99 |
| --- | ---: | ---: |
| Gate mutex wait | 0.00004 / 0.00017 ms | 0.00004 / 0.00038 ms |
| Marker read | 0.013 / 0.042 ms | 0.013 / 0.092 ms |
| Before-LSN check | <0.001 / <0.001 ms | <0.001 / <0.001 ms |
| **LibraVDB journal write** | **5.432 / 6.911 ms** | **5.429 / 6.445 ms** |
| After-LSN check | <0.001 / <0.001 ms | <0.001 / <0.001 ms |
| Journal readback | 0.976 / 1.552 ms | 0.983 / 1.533 ms |
| SQLite marker commit | 0.102 / 0.347 ms | 0.100 / 0.277 ms |
| Entire append gate | 6.539 / 8.087 ms | 6.583 / 7.649 ms |

Every accepted normal Recall had all substantive subphases and
an advanced marker. The first instrumented attempt wrongly required
positive elapsed time for an uncontended gate-mutex acquisition;
that coverage assertion rejected valid zero-wait samples. LSN and
marker-commit traces disproved the initial suspicion of skipped
durability. After correcting only that assertion, two fresh full
trials produced the table above. The invalid first attempt is not
included as an outcome.

The `Store.PutBayesianJournal` interval includes its preflight,
metadata-only LibraVDB insertion and related collection work; the
table does not prove that disk sync alone costs 5.43 ms. But even
eliminating the roughly 0.10 ms SQLite marker transaction leaves the
single-journal write above the 4 ms offer interval. The roughly
0.98 ms readback is material but its removal alone is also
insufficient. A plausible next architecture is a bounded, no-or-
minimal-dwell group commit of multiple independently validated
journals, with one durable LibraVDB transaction and one marker
transition per group. This is a hypothesis, not a rescue: the older
[v22 group journal](mmm-recall-group-journal-v22-results.md) with
8 ms dwell failed loaded latency, and a new design must preserve
every journal's as-of and duplicate/conflict checks, atomic
durability, readback/recovery evidence, writer freshness and
durable-before-ack behavior.

Two race-correctness trials also passed 128/128 writes and Recalls
without a reported data race or semantic violation; their timing
is not used above. Ordinary package tests and `go vet` passed.
All seven whole goals remain open, production untouched.

Reproduce:

```sh
EVENTFRAME_RUN_JOURNAL_PHASE_PARITY_V14=1 go test ./internal/store/libravdbstore -run '^TestResearchJournalPhaseParityV14$' -count=1 -v -timeout 5m
EVENTFRAME_RUN_PUBLISHED_RECALL_LOAD_V14=1 go test ./internal/store/libravdbstore -run '^TestResearchPublishedRecallLoadV14$' -count=1 -v -timeout 5m
EVENTFRAME_RUN_PUBLISHED_RECALL_LOAD_V14=1 EVENTFRAME_RACE_CORRECTNESS_ONLY=1 go test -race ./internal/store/libravdbstore -run '^TestResearchPublishedRecallLoadV14$' -count=1 -v -timeout 5m
go test ./internal/store/libravdbstore -count=1 -timeout 5m
go vet ./internal/store/libravdbstore
```

Git HEAD `1a7edb62b6be4031fd01ebeab8071b17303a7815`, Go
`go1.27.1 darwin/arm64`. SHA-256: protocol
`0cf20ae49025301525788dd739fc0a3e1801cebc69aaaa9fdf28d1f78da984ff`;
test `d228853db1cb4294754a3ccad87e4ef1e6f698c105d67eb0af7215a45ba1bd17`.
