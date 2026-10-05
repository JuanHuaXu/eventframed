# Transactional multi-key reads v35

Status: primitive correctness tests PASS. Wrapper integration and loaded benefit
are not measured; v33's failed 250ms age screen remains unresolved.

Added Ledger.GetBatch while leaving Get unchanged as a control. It reads caller-
ordered keys with one prepared statement inside one read transaction. Missing
records return Found=false; errors return no partial result vector. Duplicate
requests are allowed, counted toward bounds and return detached payload bytes.

Caps: 512 lookups, 8MiB encoded identity/kind bytes, independently 8MiB returned
payload bytes, and 1MiB per payload. Thus the request and response caps are not
one combined 8MiB memory claim. SQL checks stored BLOB type and length before
returning payload bytes, including corrupt oversized rows. Payload semantic
validation remains the consumer's job. Inputs must be immutable during a call.

## Verified boundaries

- Ordered single-Get parity for present/missing records and mixed record kinds.
- Exact contract-scoped identity: an alternate-contract request remains missing.
- Duplicate requests do not alias one another or mutate stored originals.
- Exact 512-key and 8MiB payload boundaries accept; excess count/bytes reject.
- Escaped encoded-key size is bounded, not merely the raw string lengths.
- A separate SQLite connection commits a new record after the first lookup;
  the current batch retains its original snapshot and the next batch sees it.
- Cancellation mid-read returns no partial results and releases the transaction.
- Pre-cancel, invalid kinds/identities, empty batches, corrupt storage types and
  oversized stored BLOBs reject.
- A panic at the test read boundary releases the transaction for the next call.

The external connection is used only in the isolated test database to exercise
SQLite snapshot behavior; it is not permission to bypass production ownership.
The package-private per-call read hook is nil at the public entry point. No
global instrumentation hook, write durability change or cross-call cache exists.

## Verification record

The expanded targeted suite passed three race repetitions, then full ledger,
research-memory and service race suites passed. Vet passed for all three:

```
go test -race ./internal/researchledger -run '^TestReadBatch' -count=3
go test -race ./internal/researchledger ./internal/researchmemory ./internal/service
go vet ./internal/researchledger ./internal/researchmemory ./internal/service
```

Source SHA-256:

- read_batch.go: `0bb6acc088cfbe6a9ef419995aee9496aca2e4944bb6c9cb80bf5f5a5db9c309`
- read_batch_test.go: `0f2c304c46fc15e6306c36cadd7e172d19137a52854e653bf2a9149b6c29225d`

## Next integration

Use bounded reads for durable admission/discard preflight and full original-record
readback. Preserve the old per-key path as a same-run control. Validate returned
length, key, kind, found/missing state and canonical payloads before relying on
results; malformed replies must not become apparent missing admissions. Keep
worker ownership, exact retry rules and commit-error handling unchanged.

The design bundles transaction reuse with statement reuse; a speedup cannot be
attributed to either separately without an ablation. No throughput, warmed
learning, feedback authority or real-agent quality claim is established by these
unit tests. Nothing was installed, deployed, published or pushed.
