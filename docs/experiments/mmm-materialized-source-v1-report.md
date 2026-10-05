# Materialized source-table results

## Outcome

No consistent admission speedup. The candidate's full preparation and append
cost roughly equals or exceeds the original in the later trials. Read times are
lower, but the candidate omits the original's corrupted-row allocation bounds
and some read checks. Do not call that a validated equivalent-contract speedup.
No integration or whole-goal success.

| Batch | Trial | Original append ms | Candidate append ms | Original read us | Candidate read us |
| --- | --- | --- | --- | --- | --- |
| 50 | 0 | .958 | .769 | 25.746 | 12.838 |
| 50 | 1 | .733 | .734 | 22.609 | 12.377 |
| 50 | 2 | .730 | .729 | 22.626 | 12.052 |
| 200 | 0 | 3.037 | 2.738 | 22.897 | 12.192 |
| 200 | 1 | 2.650 | 2.752 | 22.925 | 12.063 |
| 200 | 2 | 2.666 | 2.787 | 22.929 | 11.965 |

Means over 32 appends and 32*batch-size point reads per cell. Arm order rotates;
the first-trial difference is retained, not discarded or treated as proof of a
steady-state gain. These are microbenchmarks without serving concurrency.

## What was implemented

Test-only admission table with unique learner identity and unique canonical
source columns alongside exact original payload bytes. The private writer
derives source keys before a transaction, then checks exact retries and inserts
atomically. Source conflict under another learner key, including an alternate
JSON escape spelling, rejects the entire batch. The reader derives the source
again from original bytes and rejects inconsistency.

Canonical validation is included inside measured append time. It is not moved
outside the publication guard or omitted from accounting. Point-read lookup
keys are constructed outside both read timers; readback validation is inside.
The candidate is admission-only, with no feedback or migration implementation.

## Verification and limitations

[Contract](mmm-materialized-source-v1-contract.md),
[raw artifact](mmm-materialized-source-v1.jsonl),
[summary](mmm-materialized-source-v1-summary.json).
Raw SHA-256:
`2298c7b6f883f6ac9486fe0387d57e817e02642e47c017530ba89c2cc704c8b0`.

All 48,000 originals checked by source, sequence, key, and payload after reopen.
The summary checks 39 captured source files, unique cells, counts and timings;
second generation is byte-equal. The summarizer is postcollection, not part of
the captured precollection files. Collection passed in 2.00s (package2.235s).
Component and transaction race tests passed in 1.300s; ledger vet passed.

Tests cover changed bytes, equivalent-source conflict, late-conflict rollback,
cancellation before commit, exact retry after reopen, and source corruption.
Reopen is not process-crash testing. Concurrent owners and crash schedules have
not been tested. Raw SQL bypasses private-writer source derivation and is outside
the candidate contract. The candidate lacks the original reader's pre-materialization
size/type checks; malformed persisted rows could allocate too much before rejection.
Its canonical key parser also rejects NUL and U+FFFD identities, so full legacy
compatibility is unproven. These gaps preclude production adoption.

## Decision

Do not adopt this design as a write-latency rescue. The validation cost consumes
the expected index-write savings in this implementation. The read result merits
one bounded, equivalent-read-contract comparison before deciding whether this
representation has independent value. Any later admission optimization should
avoid repeated identity parsing only with a proof that an immutable, once-derived
source stays bound to the eventually committed original. Keep all validation,
durability and admission-validity obligations in the accounting.

All seven goals remain open; production, the whitepaper and remotes unchanged.
