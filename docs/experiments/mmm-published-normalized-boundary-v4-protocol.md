# Published-LSN normalized boundary v4: frozen component protocol

Date: 2026-10-02. Research-only, opt-in test. No production writer, reader,
schema migration, or default runtime change.

## Question

Can a declared normalization boundary make certified-LSN SQL cosine scores
agree with ordinary `Store.Search` when the caller's finite vector is
nonunit? V2 demonstrated that querying SQL on a nonunit stored vector is not
equivalent; v3 passed only for already-unit vectors.

## Fixed fixture and procedure

Use a fresh private four-dimensional, declared-payload collection, the
existing sortable availability and publication gate, and a single tenant.
Before any database write, convert every nonzero finite incoming vector to
unit length in a test-only write wrapper. Reject empty, zero, nonfinite, or
dimension-mismatched vectors. Convert the query by the same rule. The fixed
query is `(2,0,0,0)`; the appended contrast is raw `(3,4,0,0)`, raw
`(0,2,0,0)`, and raw `(-5,0,0,0)`. Retain the existing aligned genesis rows
and future sentinel. Do not tune vectors, thresholds, or ordering after runs.

Search SQL at the exact published LSN using declared payload columns,
`available_at_sort <= as_of`, and `embedding <=> unit_query AS distance`,
ordered by distance. Decode each EventFrame from that same SQL row. Convert
distance to similarity with `clamp(1-distance,-1,1)`; never use the raw SQL
`Score` as similarity. Compare the returned IDs and similarity to ordinary
`Store.Search` on the same committed state. Verify old-LSN and new-LSN body
membership across an append, and verify the new LSN survives close/reopen.

## Frozen gates

Pass only if invalid inputs are rejected before write/query; all stored
vectors in the fixture are unit length; SQL yields exactly the eligible
five records and excludes the future record; each EventFrame body agrees with
its ID, tenant, and availability; each converted similarity differs from
ordinary Search by at most `1e-5`; SQL similarity is nonincreasing; old LSN
excludes appended rows; and reopen preserves new-LSN result parity. Record
normal/race test and vet results. A failure is preserved as a negative result.

This is a single small fixture, not proof of top-k parity under ANN,
multi-tenant serving, general vector representations, loaded latency, old
nonunit-record migration, journal/snapshot atomicity, or Goal 6 completion.
