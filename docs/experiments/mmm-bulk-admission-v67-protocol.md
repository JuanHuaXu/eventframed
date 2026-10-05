# Bounded multi-row admission v67

Frozen before measurement. Candidate optimizes all-admission batches only:
transactional indexed preflight over all exact identities, byte validation of
existing and same-batch retries, insertion of new rows in original order with
VALUES chunks<=128rows, then receipt lookup by identity. No row-ID contiguity
assumption or conflict suppression. Total<=512 records and8MiB shared limits.
Mixed admission/feedback batches retain prepared ordered fallback. Commit and
FULL WAL remain unchanged; no public acknowledgment before commit.

Opt-in SourceOwner constructor; old resolved-source prepared strategy controls.
No service/default, dependency or production changes. Tests:512-row boundaries,
same-batch/old/new retries, ordered receipt/replay parity, late source conflict
after a chunk, feedback-first rejection, cancellation/panic and owner recovery.
Full ledger/memory race tests and vet before measurement.

Same isolated cost design:50/200 records,cold/64-label trained,3 trials,2 rotated
strategies,32 fresh/retry/verified-terminal cycles:24 cells. Compare full training
and original hashes, lifecycle and reopen checks. Capture raw times and source
hashes in exclusive JSONL; no competing task-started performance work.

Screen unchanged fromv66:200-record fresh mean improves>=10% in every paired
trial/state; retry mean regresses<=10% in every pair at both sizes. Integrity
mandatory. All failures retained. Passing this screen alone would not establish
mixed-service non-harm, continuous-learning accuracy or production readiness.
