# Post-guard original readback v39 results

Status: PARTIAL. Completed observations improve in all three finite trials, but
all three FAIL the unchanged 250ms age screen. No direction-level completion or
real-agent learning result follows from this cold, unlabeled storage fixture.

## Boundary and tests

Service-side validation and actual durable admission remain under the service
guard, including comparison of the owned worker's returned original with the
cold-fixture preview. Only the immutable ledger reread moves outside, followed
by the already post-guard typed discard. Every stored original field is compared.
The consumer waits for both operations before counting completion or taking a
new group. A persistence reread does not grant fresh service/history authority.

Missing, mismatched and canceled rereads prevent cleanup; a subsequent valid
retry still reaches a fresh terminal. Publication/reopen tests cover both
placements: a policy change can complete between admission and post-processing,
stale admission rejects, failed operations do not increment terminal counters,
originals survive replay and discarded records cannot become labeled feedback.
Three focused race repetitions, full ledger/learner/service race suites and vet
passed before the experiment. No ledger or learner implementation changed here.

Guard time excludes post-processing, while total duration and accepted age
include it. DurableVerifyNS on the new path measures the post-guard database
reread/comparison; the owned-return comparison remains inside CallbackNS.
Phase values therefore should not be treated as identical microbenchmarks across
placements. End-to-end completion and serving metrics retain the same meaning.

## Same-run results

[Frozen protocol](mmm-postverify-v39-protocol.md),
[raw artifact](mmm-postverify-v39.jsonl). Twelve rotated cells each execute
192 recalls with four readers and 96 future writes. All times are nearest-rank
milliseconds. Age is conditional on completed observations; dropped ones are
reported separately, not treated as fast completions.

| Trial | Path | Complete / 192 | Dropped | Age p95 | Read p99 | Write p99 |
| --- | --- | --- | --- | --- | --- | --- |
| 0 | Off | n/a | n/a | n/a | 34.009 | 17.878 |
| 0 | Non-durable group4 | 192 | 0 | 131.310 | 23.865 | 20.739 |
| 0 | Post-discard control | 163 | 29 | 338.025 | 26.929 | 33.429 |
| 0 | Post-readback + discard | 174 | 18 | 302.213 | 24.273 | 30.982 |
| 1 | Off | n/a | n/a | n/a | 32.832 | 19.804 |
| 1 | Non-durable group4 | 192 | 0 | 106.620 | 30.108 | 20.292 |
| 1 | Post-discard control | 161 | 31 | 332.935 | 24.879 | 35.956 |
| 1 | Post-readback + discard | 178 | 14 | 290.242 | 25.035 | 37.624 |
| 2 | Off | n/a | n/a | n/a | 33.988 | 17.978 |
| 2 | Non-durable group4 | 192 | 0 | 101.766 | 19.988 | 19.836 |
| 2 | Post-discard control | 159 | 33 | 341.638 | 29.511 | 37.283 |
| 2 | Post-readback + discard | 177 | 15 | 300.068 | 26.202 | 36.477 |

Completion increases from 483/576 (83.85%) to 529/576 (91.84%). The new path
persists, rereads and discards 26,450 actual original records. There are zero
recorded errors or entry expiries; the remaining 47 losses are queue drops.
All three read-p99/off <= 1.10 screens pass, but all three age screens fail.
Writer p99 improves against control in two trials and worsens in one; every
experimental writer p99 exceeds off. This is not a general serving-latency or
population reliability guarantee.

Entry totals fall from 134/132/138ms to 92/60/81ms across trials. Experimental
post-processing still consumes 300/318/319ms in total and remains included in
completion timing. Counts differ by arm, so aggregate phase sums are not
per-record costs. The result supports narrower guard scope for this typed
research lifecycle, not removal of durable work or authority checks.

## Audit and next work

The full non-race experiment passed in 21.09s. Independent parsing verified
twelve unique cells, embedded source hashes, request/write counts, group-weighted
outcome conservation, 50 validations per completion, exact durable counters and
inside/post-guard timing containment.
Artifact SHA-256:
`d4b508b63c3866629f91da4e0d1ff8c9f64f4dbfe5b737bd8a2097b129ccb4dd`.

The next performance lead is bounded transaction-scoped statement reuse in the
remaining write path, first measured in an isolated paired ledger experiment
with identical FULL commits and exact retry/corruption/crash semantics. Earlier
profiles did not establish SQL parsing as dominant, so its benefit remains an
open hypothesis; do not weaken acknowledgment or durability to force a pass.
Do not keep shifting authority-bearing admission/validation outside the guard.

This storage work does not close unique journal/event admission identity,
warm loaded learning, verified feedback or retained-history authority. Those
correctness/integration requirements remain necessary alongside performance,
and all seven research directions stay open. Prior protocols/results are
preserved. No production configuration, paper publication or push occurred.
