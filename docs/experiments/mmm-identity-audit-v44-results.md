# Durable service identity audit v44

Status: CONFIRMED missing composition guarantees. This is characterization of
uninstalled research APIs, not a production exploit or a failed SQLite atomicity
test. The test passing means the gap reproduced, not that the composition is safe.

## Evidence

[Raw artifact](mmm-identity-audit-v44.jsonl) contains eight cases: original/prepared
SQL, single/batch admission, and same-owner/reopened-owner. Each uses a real
research service journal and event from the public fixture, not fabricated
service provenance. Admission passes the existing service mutation guard and
journal/query/feature validation before the durable call.

Every case establishes the same sequence:

1. Admit learner ID 1 for one actual journal/event pair.
2. Retry ID 1: returns the original as a retry, correctly.
3. Admit ID 2 with the same binding and forecast inputs: accepted as new.
4. Change service policy: a new guarded admission correctly rejects.
5. Call raw Durable.Feedback for IDs 1 and 2: both terminals persist.
6. Replay the exclusively owned ledger: two labels are counted for the one
   service journal/event, with no pending records remaining.

As an adjacent-path control, the older in-process ResearchFeedbackBridge rejects
the same journal twice. The gap is in composing the new durable APIs into a
service-owned learner, not a failure of that bridge or exact learner-ID retries.
SQL mode and reopen do not remove it. All eight cases passed under race in
0.16s; independent parsing verified case uniqueness, embedded hashes and every
reported boolean/label count. Artifact SHA-256:
`45a204def8676014410bc8edd4efd08ddc211df92a716db048340a5747ab3daf`.

## Classification

- CONFIRMED: durable keys use tenant/stream/contract/learner ID. ServiceBinding
  is validated metadata but has no persistent uniqueness constraint or lookup.
  Per-call validation cannot deduplicate a valid journal/event across calls.
- CONFIRMED: raw durable feedback has no current service proof input; replay
  validates stored structure/order, not ongoing service/history authority.
- NOT A NEW CORE CONTRACT BUG: these low-level APIs explicitly require trusted
  callers and do not claim to authenticate usefulness or service validity. The
  missing service owner must enforce those properties before deployment.
- STILL OPEN: protecting all retained training dependencies through publication
  changes, rather than checking only the next pending record. This audit does
  not prove a complete model-history rule.

There is no corrupted production data to repair. No production or actual
private-session content was used. The storage load results v42/v43 remain what
they measured: cold original admission/readback/discard, not safe learning.

## Required rescue

Implement an opt-in durable service owner with persistent journal/event identity
and separate admission/feedback authority. A bounded in-memory tombstone set
alone cannot cover long-running streams and restarts; an unbounded map would
violate resource goals. Use an indexed durable identity keyed by owner stream,
tenant, model contract and service journal/event. Do not key by mutable features,
label value or snapshot version: changed contents under the same source identity
are conflicts, not permission to count another observation.

Identity reservation and original admission must commit atomically. An exact
source retry resolves to the original learner ID/forecast; a conflicting retry
rejects before staging. Multiple distinct journals may legitimately refer to the
same event, so event ID alone is not a deduplication key. Retain terminal identity
after labels/discards, and preserve it through reopen. Existing duplicate-bearing
logs must fail strict activation with a diagnostic, not be silently rewritten.

Fresh feedback must enter through a service-owned operation, not by exposing the
raw durable method. Verify source identity, pending/terminal status, evidence
availability, current policy/provenance and retained-history validity before
durable feedback and fitting. Explicit usefulness still needs an external
evidence contract; a signature or a correct index does not establish truth.

The next implementation stage is the bounded indexed identity primitive with
atomic retry/conflict/reopen tests, followed by its service wrapper. Keep the
unsafe composition here as the frozen control, not a behavior to preserve in the
new owner. All seven research directions remain open. Nothing was deployed or
pushed, and no whitepaper claim was upgraded.
