# Sort-key migration v1: finite backfill feasible, publication unproved

Protocol: [mmm-sort-key-migration-v1-protocol.md](mmm-sort-key-migration-v1-protocol.md).
An opt-in private-store test created an existing EventFrame collection with
no `available_at_sort` schema or row keys. The new SQL predicate initially
failed to bind. `ALTER TABLE ... ADD COLUMN available_at_sort TEXT` worked
on this vector collection. Parameterized SQL `UPDATE`s derived keys from
the **durable decoded EventFrames**, and a full-scan readiness check rejected
both the unbackfilled and one-of-three-backfilled states. After all three
rows were keyed, the check passed and exact-LSN SQL at `.120Z` returned
`{past100, at120}` while excluding future `.125Z`. Readiness and query
survived close/reopen. Deliberately corrupting one key made readiness fail.
The ordinary and race-instrumented focused tests passed.

Correction after the [receipt bug hunt](mmm-sortable-receipt-v1-results.md):
the original readiness helper could mistake an absent schema entry for
`StringField` because that enum is zero. A declared-key predicate had already
failed before ALTER and the exact-LSN query passed after ALTER/reopen, but the
helper's schema check was not valid evidence on its own. The repaired helper
now inspects the durable catalog type and verifies exact-LSN SQL binding;
the absent-schema control, migration and reopen tests pass again. The source
hash below records the original run, not this corrected helper.

Event payload bytes, stored vectors and per-version ingestion motion were
unchanged. LibraVDB record versions rose from 1 to 2 for all three rows,
and the database reached LSN 30, while eventframed's runtime snapshot did
not advance. Crucially, `ResearchSnapshotCompatible` still accepted the
pre-migration runtime snapshot after backfill. It cannot by itself certify
sort-key migration or invalidate readers during partial backfill.

**Decision:** in-place ALTER/backfill is feasible for this tiny private
collection, but it is **not a serving migration protocol**. A candidate
needs a durable pending/ready publication state, writer fencing or an
equivalent generation boundary, crash/reopen recovery, and proof that every
serving path refuses the new SQL while migration is partial. A full scan
on every read would defeat the hot path; readiness must be certified and
maintained across authorized writes. Concurrency, large-corpus cost,
sidecar authority and full service behavior remain untested. Production
was not touched.

```sh
EVENTFRAME_RUN_SORT_KEY_MIGRATION_V1=1 go test ./internal/store/libravdbstore -run '^TestResearchSortKeyMigrationV1$' -count=1 -v -timeout 2m
EVENTFRAME_RUN_SORT_KEY_MIGRATION_V1=1 go test -race ./internal/store/libravdbstore -run '^TestResearchSortKeyMigrationV1$' -count=1 -v -timeout 2m
```

Source SHA-256 at run: test `2a029a02610ebfc4d2c647e90c9a825c3f9f96e0b7b0633f0f761d4bc82f05f2`;
protocol `b936dbe1d6ddefddc4476a25bbff078214e38c5f46f31212f522770e6da23fce`.
