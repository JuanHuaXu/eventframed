# Explicit feedback bridge v1

2026-09-12. Research-only, in-process bridge from committed frontier observations
to the actual background learner. No production deployment or scoring authority.

## What was exercised

`ResearchFeedbackBridge` verifies the durable tenant-scoped journal, as-of time,
snapshot, event membership and baseline probabilities before admitting predictions.
All candidate forecasts are journaled before explicit feedback is released.
Malformed admission rolls back partial worker journals. Feedback binds the
journal/event pair, not an array position or the most recently recalled event.

The bridge retains at most 256 seen journal IDs and 256 pending candidate labels.
Seen IDs are not evicted: saturation rejects new observations rather than making
old evidence replayable. Accepted labels cannot be admitted twice. Missing labels
can be discarded without negative evidence. Invalid/early labels retain pending
state for valid retry. Worker admission enforces label-availability ordering.

One bridge is bound to one exact store snapshot. Dependency changes reject
admission, feedback, and research scoring. Score checks dependencies before and
after reading an immutable model. This is deliberately conservative: even benign
future ingestion invalidates it. It is not the completed temporal-reuse design.

## Verification

- Repeating a successful Recall with the same journal cannot create a second
  evidence set within this bridge's lifetime.
- Wrong journal/event bindings, cross-tenant journals, early labels, repeated
  feedback and changed baselines reject.
- A malformed second candidate leaves no partial prediction journal; a valid
  retry succeeds. Duplicate event membership rejects.
- Missing feedback is discarded without a negative; two delayed explicit labels
  produce exactly two completed worker updates.
- Store ingestion changes dependencies: subsequent feedback and research scores
  reject. Closed and saturated bridges reject.
- Sixteen actual service requests produce 48 explicitly labelled fixture records.
  After refits, all 512 feature forecasts equal the synchronous control exactly.
- Three repetitions of combined bridge/frontier/shadow/snapshot/background race
  tests pass. `go vet ./internal/service ./internal/researchmemory` passes.

The initial stale-dependency fixture used an idempotency key different from its
event ID and failed store validation. Correcting the fixture, not the store
contract, made the intended stale-dependency test executable. No experimental
outcome was discarded or labelled a model failure because of that setup error.

## Boundaries not established

This bridge trusts its in-process tap producer for lexical features and its
caller for externally verified usefulness. It does not authenticate truth, infer
labels, or establish independence across different but correlated journals.
Durable journals bind membership/baselines, not the research feature bits.

Deduplication is memory-resident and lifetime-bounded. Recreating a bridge loses
its replay ledger; persistence and controlled epoch transitions are prerequisites
for continuous deployment. No cross-process replay protection is claimed.

Snapshot checking does not atomically lock the database through an asynchronous
fit. Already accepted work can finish internally after a dependency change, but
the bridge refuses to expose scores under the changed snapshot. Nothing feeds
those scores into serving. Atomic publication authorization remains separate.

The tests use dummy public fixtures and an in-memory store, not real-agent
usefulness outcomes. This result establishes lifecycle behavior, not better
retrieval, calibrated confidence, or loaded end-to-end p99. Persistent concurrent
load and a viable real-task representation remain required by the roadmap.
