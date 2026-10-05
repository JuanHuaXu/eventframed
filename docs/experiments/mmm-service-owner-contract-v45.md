# Durable research service owner: implementation contract

Design only; not yet implemented or validated. Motivated by the reproduced
[v44 identity/authority gap](mmm-identity-audit-v44-results.md).

## Identity and atomicity

For one owned learner stream, use the logical key
`(tenant, stream, record contract, service journal ID, service event ID)`.
Epoch must remain consistent with the owned stream's existing replay contract;
changing epoch or snapshot cannot manufacture a second identity for the same
source journal. Different journals for the same event are distinct observations.

The durable original supplies the learner ID and complete pre-outcome experts.
A source-key lookup must be indexed and bounded in returned records/bytes, not
a whole-ledger scan on each call or an indefinitely growing RAM map. Pending
requests keep the existing count/byte limits. Persistent tombstones may grow on
disk; checkpoint/retention is a separate unresolved requirement, not solved by
calling the index bounded.

Reserve source identity and original record in the same SQLite transaction.
No acknowledgment before FULL commit. Resolve exact retries to the same learner
ID and original bytes. Validate immutable forecast parameters and binding on
retry; differing inputs are conflicts. Any uncertain commit stops the owner
until replay, preserving the existing fail-closed recovery contract. The caller
must not allocate a fresh identity merely because a reply was lost.

Strict activation over an existing log must verify uniqueness. Duplicate-bearing
history is rejected diagnostically; never silently merge, delete or reinterpret
its labels. No live/default migration is authorized by this design document.

## Authority boundaries

Admission requires current guarded service journal/query/feature validation.
The index establishes identity, not provenance validity. Label admission requires
a distinct explicit evidence contract and current service/history checks while
crossing its persistence boundary. Missing feedback is neither negative nor
permission to forget identity. A terminal discard stays unlabeled and final.

Model publication must account for the dependencies of retained training history,
not solely the current pending record. Failing history validation suppresses use
of the affected learner until an explicit reset/rebuild policy applies. Replay
alone never grants fresh evidence or serving authority. Do not publish a general
history claim until the concrete dependency representation and tests exist.

## Falsifiers and tests

- Same source/new caller ID cannot create a second admitted/learned record.
- Exact retry resolves to the identical original before/after reopen.
- Different journal/same event and different event/same journal remain distinct.
- Mixed new/retry/conflicting batches are all-or-nothing at the durable boundary.
- Cancellation, partial staging, lost acknowledgments and process exits on both
  sides of commit preserve recovery, identity and original bytes.
- Changed policy rejects feedback through the service owner even when the raw
  storage API could accept it; retained-history invalidation is tested separately.
- No label after a terminal discard; no missing-label-to-false conversion.
- Lookup work is measured across increasing historical record counts and cannot
  hide a linear scan or unbounded in-memory deduplication cache.

Only after those tests should the guarded load path use this owner, with the
old unsafe composition retained as a clearly labeled research control. Keep the
existing completion-age, serving and durable integrity screens unchanged.
