# Envelope storage feasibility result

## Decision

Promising storage lead, not a Goal 6 rescue. A single envelope reduces measured
append work relative to one row per event. A small second marker is cheap enough
to justify investigating a guarded activation protocol. Neither form currently
implements the source-identity, replay or admission semantics of the daemon.

[Frozen contract](mmm-admission-envelope-v1-contract.md),
[raw run](mmm-admission-envelope-v1.jsonl),
[summary](mmm-admission-envelope-v1-summary.json),
[summary replay](mmm-admission-envelope-v1-summary-replay.json).

Raw SHA-256:
`08aa9d86b34799d161f6897ed4c039c433bf53ec314d83c4d7203e10bbeca02a`.
All 28 captured source hashes match embedded and local files. Summary replay is
byte-identical; replay is not a second independent experiment.

## Results

Three rotated trials, 32 batches per cell, sizes 50 and 200, three modes:
18 cells and 576 batches. Identical ordered 1,024-byte synthetic event payloads
are retained in each encoding. FULL/WAL is asserted in every cell. Every stored
row is checked byte-for-byte with its sequence after reopen: 24,576 rows across
all cells. All cells also check exact retry and rejection of conflicting bytes.

Trial-mean ranges, milliseconds:

| Events | Per-event append | Envelope append | Staged preparation + marker | Marker alone |
| --- | --- | --- | --- | --- |
| 50 | 0.457-0.586 | 0.186-0.208 | 0.262-0.288 | 0.075-0.082 |
| 200 | 1.676-1.696 | 0.555-0.573 | 0.604-0.795 | 0.063-0.078 |

Paired envelope/row mean ratios are 0.356-0.420 for 50 events and 0.331-0.338
for 200. Staged total/row ratios are 0.492-0.586 and 0.356-0.474 respectively.
Marker-only ratios are 0.140-0.172 and 0.037-0.047, but these are not total work
or attainable daemon speedup ratios. The raw artifact retains each batch's
timings; means and p95s are reported separately, including outliers.

The full ledger race suite passed (3.917 s), vet passed, and the isolated
non-race experiment passed (0.82 s test, 1.044 s package). No concurrent
task-started experiments ran. This test changes no ledger implementation.

## What is and is not measured

Measured intervals cover calls to unchanged `AppendBatchPrepared`, including
its validation, SQL and commit. Preparation time is included in staged totals.
Encoding and digest calculation, event validation, worker staging, source
lookup, reopening and readback lie outside those intervals. There is no actual
service guard, policy change or outcome-bearing learning in this experiment.

The common schema validates generic ledger structure only. It does not install
the real per-source uniqueness index. Envelope storage removes per-event SQL
rows, not the obligation to enforce each event's identity. The staged marker is
a generic row in a separate namespace, not an implemented acceptance operation.
Hash equality is byte binding, not evidence authentication or policy validity.
Reopen tests are not power-loss tests. These are short, consumed, synthetic
microbenchmarks without confidence bounds, not validated serving claims.

## Next discriminating experiment

Before service integration, restore event-level identity enforcement in an
isolated envelope prototype: store the immutable payload once and maintain
unique source-to-envelope/offset references. Compare one-transaction envelope
acceptance against per-event control with the same identities and retry rules.
This simpler form avoids adding a prepare/activate lifecycle initially. If the
per-source references erase the speed advantage, do not claim the marker result
survives that missing work.

Required rejection cases: duplicate source within/across batches, conflicting
retry bytes, mixed old/new sources, late invalid members, malformed offsets,
missing envelope, and failure before/after commit. Replay must return the exact
original per event, not a fresh forecast. No partial acknowledgement is allowed.

Only after this can staged preparation be assessed: the acceptance marker must
be committed while a compatible service guard is held, and unactivated prepared
payloads must never be read as admitted evidence. Crash/retry state, worker IDs,
pending capacity and source reservations need an explicit contract. The previous
interleaving counterexample still rules out an unguarded final recheck.

Production, whitepaper, credentials and remotes remain untouched. All seven
goals remain OPEN; no existing performance or quality failure is reclassified.
