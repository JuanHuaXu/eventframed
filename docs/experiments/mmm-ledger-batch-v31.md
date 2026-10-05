# Atomic ledger batch primitive v31

Status: structural transaction tests PASS; wrapper integration and loaded
performance remain untested. The durable v29/v30 failure is not yet rescued.

Added research-only Ledger.AppendBatch, leaving the existing Append implementation
unchanged as an independent control. At most 512 entries and 8MiB of encoded
identities, kinds and payloads are admitted per transaction. Existing per-field
4096-byte and per-payload 1MiB bounds still apply. Batch inputs must be immutable
until return. Cancellation is checked before and during validation and execution.

Records execute in input order. Exact retries retain their sequences, including
repeats within a batch. Conflicting bytes roll back the entire transaction. A
terminal requires the same identity's prior admission, either already committed
or inserted earlier in this transaction. All acknowledgments wait for COMMIT;
errors return no result vector. A commit error can be uncertain and must be
resolved by replay/retry, not assumed to mean rollback. WAL/FULL synchronous
durability and local ownership remain unchanged.

## Verified boundaries

- Sequential single-Append replay/sequence parity and concurrent exact retries.
- Mixed retries/new records and same-batch duplicate identity idempotence.
- A late conflicting terminal rolls back preceding new inserts.
- Terminal-before-admission rejection, malformed payloads and invalid identities.
- Empty/oversized batches, exact 512-entry/8MiB acceptance and one-byte excess.
- Cancellation before execution and after all inserts but before COMMIT.
- Panic before COMMIT rolls back and releases the connection for the next call.
- Actual child-process exits before and after COMMIT through the batch API,
  with zero-or-complete recovery, original-byte comparison, exact retries and
  SQLite integrity checks after reopening.

The package-private, per-call pre-COMMIT hook exists only for deterministic fault
tests; the public production entry always passes nil. No shared global hook or
SQL durability override was introduced. Process termination is not hardware
power-loss validation and a SQLite transaction is not a cross-store transaction.
The ledger does not validate evidence truth or forecast payload semantics.

## Verification record

Targeted batch tests passed three race repetitions, then the expanded suite
(including panic rollback) passed in the full ledger/research-memory/service
race run. Vet passed for all three packages. Commands:

```
go test -race ./internal/researchledger -run '^TestBatch' -count=3
go test -race ./internal/researchledger ./internal/researchmemory ./internal/service
go vet ./internal/researchledger ./internal/researchmemory ./internal/service
```

Source SHA-256 at this checkpoint:

- batch.go: `addda504291854fe1a63770315ef7e468ed3ee258e6d14c972bd4da0c1fbab5f`
- batch_test.go: `d2e24293243859821baf4c9cf47c99b089c3498b18d794404b3967c5fff8f318`
- batch_crash_test.go: `d2427050d0eb664b32a60da08e89c9a21f2bef7b7357adf08dc7a8ea304e014d`

## Next boundary

Add separate bounded batch-admission and typed batch-discard operations to the
exclusively owned durable wrapper. Unsupported log implementations must reject
before worker mutation. Preserve original forecasts and pending identity, stop
the wrapper on uncertain commit, and test reopen/replay and exact retry across
failure boundaries. A failed batch cannot silently leave an operational learner
ahead of its durable history. Do not generalize the prior cold-preview ID mapping
to a concurrently learning model.

After wrapper tests pass, rerun the same grouped load with individual transactions
as control and integrity readback retained. No daemon configuration, serving
policy, published paper claim, deployment or push changed at this checkpoint.
