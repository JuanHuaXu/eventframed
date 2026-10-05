# Normalized envelope: identity-restored test

## Decision

Do not promote. Restoring unique per-event references consumes most of the
earlier envelope advantage, and source lookup regresses. This is a useful
negative result for the initial storage-only interpretation, not an admission
or serving rescue. All seven goals remain open.

[Contract](mmm-envelope-reference-v1-contract.md),
[raw data](mmm-envelope-reference-v1.jsonl),
[summary](mmm-envelope-reference-v1-summary.json),
[replay](mmm-envelope-reference-v1-summary-replay.json).
Raw SHA-256:
`b3038f9799e08f4fbeeaedd55c386c0e58fb95c67aa8bdac48f571866bef6d88`.
All 30 captured source hashes match embedded/local files; summary replay is
byte-identical. This is analysis replay, not fresh confirmation.

## Implementation and checks

The entire candidate lives in test files. It stores one immutable payload blob
per fresh batch and a normalized row per event containing unique original key,
unique service-source identity, stable sequence, envelope ID, byte start and
byte length. The body and all references commit in one transaction. Exact retries
retain their original bytes and sequence; a conflicting source rolls back the
whole batch. Missing bodies and invalid slice bounds reject rather than becoming
cache misses. No unguarded acceptance or prepare/activate split was introduced.

Unit tests cover duplicate sources, conflicting retries, mixed old/new records,
late invalid members, rollback after insertion, reopen, lost acknowledgment,
invalid offsets and missing bodies. The full ledger race suite passes (4.001 s),
as does vet. These are not power-loss tests. No daemon constructor installs the
schema and no current ledger implementation was edited.

The unchanged control enables its real service identity unique index. Both arms
use FULL/WAL, identical synthetic bound payloads and one transaction per batch.
Candidate append timing includes its parsing, identity encoding and body assembly.
Twelve rotated cells cover sizes 50/200 with 32 batches over three trials. All
48,000 originals were verified by source after reopen, including exact bytes,
sequence and source identity; first-batch retries also retain sequences. The
experiment passes in 2.64 s (2.861 s package).

## Results

Trial mean ranges; append in milliseconds, per-event lookup in microseconds.

| Events | Indexed row append | Envelope/reference append | Row lookup | Envelope lookup |
| --- | --- | --- | --- | --- |
| 50 | 0.718-0.982 | 0.741-0.757 | 22.82-25.50 | 25.28-25.73 |
| 200 | 2.632-2.653 | 2.294-2.299 | 23.01-23.27 | 47.61-48.06 |

At 200 events, paired append ratios are 0.867-0.872 (about 13% lower), while
lookup ratios are 2.046-2.075. At 50 events, append ratios are 0.771, 1.032 and
1.032: no consistent win. The earlier roughly threefold storage-only gain does
not survive this identity-restored prototype. The measured append benefit is
not a serving-latency claim and must not hide the read regression.

Lookup paths are not equally hardened: the control validates returned JSON
binding fields, while the prototype checks the indexed source, slice bounds
and test readback bytes. Thus even the slower candidate lookup does less
defensive validation. No security or full-contract equivalence is claimed.
Feedback, learner state, service guards, complete history replay, index migration,
process-crash recovery and long-duration scaling remain unimplemented here.

## Next falsifiable lead

The benchmark issues one source lookup per event, and the prototype's SQL slices
the shared blob separately each time. Repeated blob access is a hypothesis for
the size-dependent lookup regression, not a proven attribution. A bounded batch
lookup can resolve references first and load each required envelope once, then
validate every original and requested source. Compare against the existing
transactional source-batch API, not only the point-read control. Cap aggregate
bytes, preserve order/missing results, reject bad offsets and source mismatches,
and include complete parsing in timing. Do not hide the cost in an unbounded cache.

Only if full append-plus-read work improves should service integration be
considered. Source uniqueness and original-forecast identity remain mandatory.
Production, whitepaper, remote repositories and private data were untouched.
