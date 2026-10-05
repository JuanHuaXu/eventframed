# Durable wrapper batches v32

Status: wrapper transaction/recovery tests PASS. Loaded performance not measured;
the v29/v30 latency failure remains unresolved until the next load experiment.

Added separate AdmitBatch and DiscardBatch operations, each bounded to 256 unique
IDs. The wrapper requires atomic batch support before any worker mutation. All
inputs are preflighted, including original-byte retry contracts, binding identity,
new-ID sequence and pending capacity. Individual APIs remain unchanged.

AdmitBatch stages forecasts through the owned worker's actual Predict/Record
path, then stores those original records atomically. A batch is not promised one
shared model snapshot: asynchronous fitting may publish between predictions;
each record preserves the snapshot's actual expert outputs. No cold-preview ID
mapping is reused. Returned bindings are detached from caller input and cannot
mutate stored originals. The API does not certify service/evidence provenance.

DiscardBatch writes explicit unlabeled terminals atomically before releasing
pending records. It neither fits the model nor advances its evidence clock.
Exact retries preserve terminal dates; labels and discards are incompatible.

The owner marks itself stopped before staging or uncertain durable work. Any
failure or panic after that point leaves it stopped until close/replay. Only a
complete, consistent COMMIT acknowledgment followed by local completion restores
operation. A known SQL rollback is not permission to continue a learner whose
prediction-ID sequence may already have advanced. Preflight rejections leave the
unchanged owner usable. Neither API makes SQLite and LibraVDB one transaction.

## Verified cases

- 256-record discard, reopen and complete original preservation without learning.
- Admission after 64 processed labels: 128 warm original forecasts match owned
  records, persisted records and reopened retries, without recomputation.
- Before/after-commit injected errors and panics stop the owner; replay resolves
  absent versus committed batches and retries do not duplicate records.
- Four actual child-process exits through the durable APIs around the append
  boundary: admission/discard, uncommitted/committed, followed by replay and retry.
- Unsupported logs, invalid/duplicate IDs, late gaps, malformed probabilities,
  invalid time, oversized/empty batches and cancellation reject before staging.
- Labeled/missing/early discard members reject without discarding a valid sibling.
- Binding input/output mutation cannot alter stored originals.
- A late conflicting retry does not stage an earlier new request; mixed retry/new
  admissions and discards preserve the expected flags.
- Exact 256-pending capacity accepts; the next new prediction rejects without
  changing the owner or its next ID.

Process exit is not hardware power-loss validation. Wrapper fault injection
tests its pre-append and post-COMMIT/pre-local-completion boundaries; v31 separately
tested actual exits after SQL inserts but before COMMIT. No outcome authenticity
or complete service-history certificate follows from these structural tests.

## Verification record

Targeted durable-batch tests passed three race repetitions. The expanded suite
then passed full research-memory/ledger/service race tests and vet:

```
go test -race ./internal/researchmemory -run '^TestDurableBatch' -count=3
go test -race ./internal/researchmemory ./internal/researchledger ./internal/service
go vet ./internal/researchmemory ./internal/researchledger ./internal/service
```

Implementation SHA-256:

- durable_batch_admit.go: `bc46cb07cb161483ba2eb19390b75166a0bc7068cd7bb417852bb2a2b4507b4a`
- durable_batch_discard.go: `875aea14a5cdce68b96a7c1064b4447816b4ea8ea237407e8dcbd2ae6539e5a8`

Tests at this checkpoint:

- durable_batch_admit_test.go: `dd5062511c1abde009b2861d3d6ee26baa7e2a02da5578d49a8b62db1a286ccc`
- durable_batch_discard_test.go: `091c284aa46847eb9cf03aa200be8ce57a1b04fcb649156d7922671630ced054`
- durable_batch_crash_test.go: `9d96952e0830ec4b21bb78b4628c8003c9d6dee2f0bcd35367c1f6e1c2243f7c`

Next integrate these APIs into the same bounded group4 fixture. Retain off,
non-durable and individual-transaction controls, complete original-record
readback, actual discard writes, the 250ms age screen and write-tail reporting.
Only that comparison can establish whether batching rescues durable throughput.
All direction-level learning/real-task criteria remain open. No production
configuration, deployment, publication or push occurred.
