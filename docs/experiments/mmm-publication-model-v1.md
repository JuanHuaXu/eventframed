# Coherent publication model v1

## Audit boundary

The persistent store currently serializes mutation paths with writeMu and reads
Snapshot under its read lock. Ordinary ingestion changes RuntimeVersion and
EvidenceEpoch. Composition also changes AbstractionVersion and must NOT inherit
an ingestion-only exemption. Policy/certificate, posterior/outcome, agency,
graph/snap, deletion and other invalidation paths require general-mutation handling.
Initialization/recovery loads a durable version before the store is returned.

A transaction may become durable before the in-memory version is assigned. A
cached version reader cannot simply pretend that interval does not exist.
The Put error-resolution path can report a matching durable record as duplicate;
before a new publication mechanism treats such a path as unchanged, authoritative
transaction/state resolution must establish what happened. This is an integration
audit requirement, not a confirmed backend bug or permission to rewrite recovery.

## Executable candidate

`internal/researchpublication` is a standalone research state machine, NOT wired
to the store. It atomically publishes immutable (snapshot, bounded motion history,
pending-write declaration, quarantine) state. Readers take no writer mutex.

- Begin occurs before any backend-visible mutation and issues a unique token.
- Unknown/general pending writes reject compatibility.
- Pending ingestion may preserve only strictly earlier query views, under the
  explicit contract that this write changes nothing except that future event
  and its ingestion version counters.
- Commit validates the declared transition. Ingestion must increment exactly
  runtime/evidence counters by one and change no other snapshot component.
- An unexpected committed transition quarantines the view, not a fabricated
  rollback. A proven noncommit can abort; an ambiguous commit quarantines.
- Motion maps are copied before publication, bounded to4096 versions. Evicted
  proof history rejects reuse. Overflow, zero cutoffs and invalid tokens reject.

The reference snapshot is not a permission to publish a model. A real adapter
would still perform as-of and feedback-time checks and obtain publication
authority at the proper boundary. Each caller's writer declaration must be true;
the state machine cannot infer it from an event label.

## Tests and result

Three race-test repetitions and vet passed. Tests enumerate general/ingestion
writes before/at/after query cutoff, committed/current compatibility, duplicate
and wrong commit tokens, certain/uncertain abort, mismatch quarantine, concurrent
readers across100 commits, expiration beyond4096 commits, zero time and overflow.

This is a logic pass only. It neither fixes the measured store-lock contention
nor establishes loaded latency. Copying up to4096 motion entries per publication
has a write cost that remains unbenchmarked. The persistent store and existing
serving snapshot getter remain unchanged.

## Required before integration

Every actual writer must enter the protocol before data visibility; no exception
for composition, semantic updates, deletes or recovery. Pending-commit visibility
and uncertain outcomes need fault-injected tests. Only then test the same short
and long persistent workloads, with unchanged completion and age gates. No
shorter benchmark can stand in for that integration proof.
