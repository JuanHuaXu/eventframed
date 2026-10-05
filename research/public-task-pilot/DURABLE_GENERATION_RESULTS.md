# Durable generation coordinator: failure-boundary tests

Added internal/researchindex/durable.go and durable_test.go. This is a research
coordinator exercised against real synchronous libravdb transactions, not a
daemon replacement. A flat vector collection and metadata-only revision
collection form the test's atomic durable state. Flat is used to isolate the
commit boundary, not as an asserted scalable ANN solution.

## What is connected

Apply prepares an immutable candidate, calls a synchronous persistence callback
for records plus revision, then publishes the prepared generation on success.
Current-view acquisition is excluded during that operation. Prior views remain
valid historical states. Any callback error or panic quarantines the coordinator;
discarding the unpublished candidate is explicitly not a storage-rollback claim.
Recovery constructs a new coordinator from coherent authoritative durable state.

The callback receives copied prepared values and must not mutate/retain them.
It must return nil only after atomic durable completion. This is a trusted local
adapter contract, not a property inferred from an arbitrary callback's return.
The coordinator assumes sole mutation ownership, not concurrent external writers.

## Tests with real storage

Each case starts with one committed record, then attempts to delete it and add
another in one transaction with the revision update:

- Error before the transaction: reopen sees the original record/revision.
- Error after staging changes but before committing: reopen sees original state.
- Error injected after successful durable commit: reopen sees the replacement
  and incremented revision, despite the caller receiving an error.
- Caller canceled after successful durable commit: publication still completes.
- Ordinary success: replacement and revision agree before and after reopen.

Unknown cases block fresh views and further writes until recovery. Old immutable
views remain unchanged. Recovery uses known fixture IDs and the durable revision;
it does not claim a general snapshot-enumeration or WAL-replay implementation.
Recovered coordinators successfully accept another transaction. Additional tests
cover read-deadline expiry while a write is unsettled and panic quarantine.
The full researchindex suite passes five race-enabled repetitions:
`go test -race ./internal/researchindex -count=5 -timeout=120s`.

## Limits and next step

No low-level torn-write or fsync fault injection; the post-commit fault models
lost acknowledgement using a real completed transaction. There is no process-
crash test, durable idempotency journal, production snapshot schema integration,
HNSW builder, or service benchmark. Caller retry IDs still need the existing
EventFrame idempotency contract; Apply alone is not an idempotent public API.

The coordinator currently blocks current views through the durable call. That
is a correctness reference and could add contention; it is not a performance
claim. Next integrate immutable ANN base construction and coordinator-managed
compaction, then test actual service semantics before measuring throughput.
All seven whole goals remain open; production and dependencies are unchanged.
