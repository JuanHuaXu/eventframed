# LSN availability-schema v1: component pass

Protocol: [mmm-lsn-availability-schema-v1-protocol.md](mmm-lsn-availability-schema-v1-protocol.md).
The opt-in test-only private collection declared `available_at: StringField`
with the same vector dimension, metric, HNSW and memory-mapping options as
the ordinary store. The unchanged research batch writer stored past, future
and then visible EventFrames. Parameterized SQL with both `AS OF LSN` and
`WHERE available_at <= $available_by` returned exactly `{past}` at old LSN
20 and `{past, visible}` at new LSN 24. The future event did not leak. The
ordinary `Store.Search` agreed at the new state, the SQL result survived
close/reopen, and active temporal leases returned to zero.

The focused ordinary and race-instrumented runs passed. This falsifies the
idea that the installed LibraVDB SQL engine cannot filter availability at
all. The v1 rolling failure came from the **current collection's missing
metadata declaration**, not from unsupported SQL syntax. It does not prove
that an existing collection can be migrated safely, nor that an LSN pin can
meet loaded freshness or latency gates. Production remains unchanged.

Reproduce with:

```sh
EVENTFRAME_RUN_LSN_AVAILABILITY_SCHEMA_V1=1 go test ./internal/store/libravdbstore -run '^TestResearchLSNAvailabilitySchemaV1$' -count=1 -v -timeout 2m
EVENTFRAME_RUN_LSN_AVAILABILITY_SCHEMA_V1=1 go test -race ./internal/store/libravdbstore -run '^TestResearchLSNAvailabilitySchemaV1$' -count=1 -v -timeout 2m
```

Source SHA-256 at run: test `5e43981917a3f5e06e2a7d1cd6ce6621ff8bcde7fef9edae900684ded732aee7`;
protocol `1f77f4cf40a9e539e1bf7abd5737a97a86a6ee831e396bff85984517567ecc97`.
