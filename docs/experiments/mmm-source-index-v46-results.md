# Source identity primitive v46 results

Status: primitive correctness and indexed-lookup diagnostic PASS. The durable
service owner is not integrated yet; v44's unsafe composition remains possible
through unchanged default APIs. No safe-learning or direction-level completion.

## Persistent constraint

`EnableServiceIdentity` explicitly opts an exclusively owned ledger into a unique
expression index on tenant/stream/record-contract/service-journal/service-event.
The source identity is derived from the bound original row, so original insertion
and identity reservation share one SQLite transaction. No separate reservation
record can commit first. Changed snapshot/forecast data cannot create another
identity for the same source key. Different journals, events, tenants, streams
and contracts remain separate.

Activation checks the structure of existing bound identities, validates any
existing index definition, and refuses duplicate/invalid history without rewriting
it. It may scan/build an O(N) index at startup. The index persists across reopen
and terminal records; persistent index/tombstone storage grows with history.
No default constructor enables it, and no production database was migrated.

`GetServiceAdmission` uses all five equality fields to return the persisted
original. A missing index is an error, not a miss. Keys are bounded and valid
UTF-8; SQL bounds identity/payload materialization, with payload <=1MiB. Missing
sources return sql.ErrNoRows; canceled/ambiguous/malformed reads are not misses.
It does not reconstruct a lifetime Go deduplication map.

This is a structural index for owned bound records, not full forecast validation
or a generic JSON authenticity mechanism. The future owner must still require
canonical originals, validate binding/forecast semantics and reject unbound
service operations. Index membership does not authenticate usefulness, authorize
feedback or establish current model-history validity.

## Correctness evidence

- Single, batch and prepared-batch duplicate sources reject under a new learner
  ID. A late source conflict rolls back earlier inserts in the same batch.
- Exact learner-ID retries preserve originals; the source resolves after reopen
  and after a discard terminal. The index cannot be lost by an ordinary reopen.
- Each source-key component is tested for separation; eight concurrent claims
  for the same source produce exactly one admitted row.
- Duplicate/invalid existing history, wrong index definitions and canceled
  activation reject without changing original rows or leaving a partial index.
- Actual process exits before/after COMMIT yield absent/present identity with
  matching original retry semantics. These are not hardware power-loss tests.
- Missing-index, key-bound/UTF-8, cancellation and oversized-payload readback
  reject. The oversized blob is deliberately injected outside normal append
  validation to exercise the defensive read envelope.

Initial identity/crash tests passed three race repetitions. Full ledger, learner
and service race suites including the read-bound tests passed, as did vet.

## Lookup scaling diagnostic

[Frozen protocol](mmm-source-index-v46-protocol.md),
[raw artifact](mmm-source-index-v46.jsonl). One fresh history per size, 128 hits
and 128 misses each, no excluded samples. Payloads are storage fixtures, not
learned predictions. Times below are microseconds; p95 uses nearest rank.

| History records | Hit mean | Hit p95 | Miss mean | Miss p95 |
| --- | --- | --- | --- | --- |
| 100 | 45.808 | 62.666 | 41.335 | 47.583 |
| 1,000 | 29.438 | 31.333 | 26.444 | 28.000 |
| 10,000 | 26.211 | 28.208 | 22.557 | 23.291 |

Every query plan reports `SEARCH research_log USING INDEX
research_service_identity_v1` with all five expression equalities, not a full
history scan. Larger sizes running faster in this ordered diagnostic can reflect
warmup/cache noise; it is not a scaling speedup or a population tail guarantee.
Index construction, replay, concurrent writes and billion-record behavior were
not benchmarked by these point lookups. The result supports proceeding to the
owner API, not claiming its loaded overhead is negligible.

The opt-in experiment passed in 0.16s. Independent parsing verified sizes,
128+128 counts, embedded source hashes and the five-field search plans.
Artifact SHA-256:
`4892794512d9ee513a7d79580e534e6f71b74e96d0d46f249b88b8ef1e7ac5eb`.

## Next step

Build the opt-in service owner that resolves source keys before staging learner
IDs, validates immutable retry parameters, and returns the original on exact
source retry. Keep its raw durable object private. Strict activation must also
perform existing canonical replay validation. Re-run v44 against that owner,
with old APIs retained as the unsafe-composition control. Fresh feedback and
retained-history authority remain separate necessary work. All seven directions
remain open; nothing was deployed, published or pushed.
