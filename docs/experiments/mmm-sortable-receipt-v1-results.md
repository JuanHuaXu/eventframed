# Sortable receipt v1: exact boundary, isolated cost pass

The frozen [functional](mmm-sortable-receipt-v1-protocol.md) and
[cost](mmm-sortable-receipt-cost-v1-protocol.md) screens pass for an
opt-in, research-only EventFrame batch writer. This is a Goal 6 prerequisite,
not marker renewal or production readiness.

## Functional result

`PutResearchEventBatchSortableReceipt` commits the event rows and runtime
state in one LibraVDB transaction and returns its exact commit LSN. A
two-event batch returned a positive LSN equal to `LatestCommitLSN`, advanced
the runtime version and evidence epoch by two, preserved per-event motion,
and stored fixed-width UTC keys. Exact-LSN as-of queries excluded a future
subsecond event, including after close/reopen. An exact retry returned
duplicate results and zero LSN with no new commit. A mixed duplicate/new
batch advanced once and returned the new boundary. Conflicting duplicates
and an existing unkeyed legacy row failed without new commits.

The bug hunt found that the initial sortable writer and migration readiness
helper were using `MetadataSchema["available_at_sort"] == StringField` as a
declaration test. `StringField` is enum zero: an absent map entry passes that
comparison. An undeclared-schema writer control and a keyed row with no SQL
declaration failed before repair. A second falsifier showed that a numeric
column can bind the string predicate by coercion, so binding alone does not
prove a text key. All three controls now pass. The final check verifies the
column type in LibraVDB's durable catalog snapshot and then binds an
exact-LSN, zero-row reader predicate. It assumes the EventFrame collection's
`id` column is the native text key and requires the sort key to have that
same catalog type. LibraVDB deliberately stores SQL declarations in its
catalog while reopened physical `Collection.Config()` omits relational
columns; the initial map-presence repair wrongly rejected valid migrations.

A `pg_class`/`pg_attribute` type-check attempt was rejected after a private
`ALTER TABLE` produced relation OIDs 103 and 105 for the table and new
attribute respectively. The actual as-of query still worked. The result
does not treat those virtual views as a reliable join for this migration.

The private migration and publication probes pass again with the catalog-
bound check. The undeclared-schema writer control checks not just the API
error but unchanged LSN/runtime state and absence of the event record.
Focused normal and race tests, ordinary package tests, and package vet pass.

## Isolated writer cost

Three rotated pairs each used fresh private collections, 16 warm-up and 128
measured one-event commits per arm. Both arms performed a post-call LSN
read for correctness; that read was excluded from the timed call. All 864
writes were new and final row/version counts matched. Ratios are receipt
writer divided by existing sortable writer; the frozen screen required
every block p50 <=1.10 and p99 <=1.20.

| Block | Control p50 | Receipt p50 | p50 ratio | Control p99 | Receipt p99 | p99 ratio |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 6.978 ms | 6.991 ms | 1.002 | 8.544 ms | 9.033 ms | 1.057 |
| 1 | 6.977 ms | 7.019 ms | 1.006 | 9.045 ms | 8.959 ms | 0.991 |
| 2 | 7.017 ms | 6.967 ms | 0.993 | 9.066 ms | 8.914 ms | 0.983 |
| Pooled | 6.985 ms | 6.995 ms | 1.001 | 9.066 ms | 9.003 ms | 0.993 |

These are noisy sequential private-store timings, not confidence intervals
or loaded service latency. Earlier schema-check variants also passed their
cost screens, but the table above is the final catalog-type-checked path.
The receipt adds no
detectable large cost here; that conclusion does not include the SQLite
marker transaction, cross-process fencing, crash reconciliation or Recall.

## Reproducibility and limits

- Functional: `go test ./internal/store/libravdbstore -run '^TestResearchSortable(EventBatchReceipt|EventBatchRequiresDeclaredKey|EventBatchRejectsWrongKeyType|ReadyRequiresDeclaredSchema)$' -count=1`; the focused migration/publication tests require their existing opt-in environment flags. The same combined focused suite passed under `-race`.
- Cost: `EVENTFRAME_RUN_SORTABLE_RECEIPT_COST_V1=1 go test ./internal/store/libravdbstore -run '^TestResearchSortableReceiptCostV1$' -count=1 -v`.
- Protocol SHA256: `01a258b958c5e1271dc1e1337d5e8155c52776f6ef8b1d107f4e790880c0ecb2` and `52d05053133766d582d8310ca245868de2e4b6316f7d635646ae96164f056c29`.
- Writer SHA256: `2f5607e43a40eec7d0ef892114f153e7df413f0202cad9bd0e53a71202121b09`.

A receipt identifies a committed boundary, but cannot prove that another
process did not write between two boundaries. A safe incrementally renewed
READY marker still needs a writer fence or equivalent complete transaction
accounting, a recoverable cross-store handshake and loaded freshness tests.
All seven whole goals remain open; production is untouched.
