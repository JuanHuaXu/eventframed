# Prepared durable wrapper load v41 results

Status: MIXED/PARTIAL. The isolated v40 benefit does not establish the loaded
250ms age target: all three prepared trials FAIL it. Keep the constructor opt-in.

## Wiring and correctness

`OpenDurablePreparedBatches` delegates opening/replay to the existing constructor,
then replaces its private batch-append method before the owner is exposed.
Single-record writes, reads, replay, ownership, FULL commits and acknowledgments
are unchanged. Both loaded durable arms use group4postverify; PreparedWrites
distinguishes their SQL implementation. The default constructor is untouched.

An initial warm parity test FAILED intermittently because it did not control
forecast-producing training history. Waiting for the final label count alone
does not match intermediate model publications: mixture updates use original
pre-outcome experts. A diagnostic readback changed the scheduling and passed,
which was not accepted as a repair. A deterministic control subsequently held
fitting until all 64 originals were recorded, versus waiting for each update,
using the same constructor and labels; the final original was cold in the
delayed schedule and warm in the sequential schedule. Thus asynchronous schedules
are not valid interchangeable histories for this SQL parity comparison.

The corrected parity test waits for each training update and checks the actual
persisted training originals across owners. It then compares 200 warm admission
records, original readback, reopen/admission retries and terminal reopen retries.
These and uncertain-commit tests before/after actual commit pass three race
repetitions for both constructors. Small prepared-load accounting also passes
three race repetitions. Full ledger/learner/service race suites and vet pass.
This controlled test does not establish schedule-invariant online learning or
warm loaded learning. The earlier failed test attempts are retained in this
account, not reclassified as successful runs.

## Same-run experiment

[Frozen protocol](mmm-prepared-load-v41-protocol.md),
[raw artifact](mmm-prepared-load-v41.jsonl). Twelve rotated cells, each with
192 recalls/four readers, 96 future writes, 50 candidates and the unchanged
bounded grouping/queue/guard settings. No labels or fitting in load cells.
All times below are nearest-rank milliseconds. Completion age includes original
readback and durable discard; dropped observations are not assigned fast ages.

| Trial | Path | Complete / 192 | Dropped | Age p95 | Read p99 | Write p99 |
| --- | --- | --- | --- | --- | --- | --- |
| 0 | Off | n/a | n/a | n/a | 33.105 | 18.704 |
| 0 | Non-durable group4 | 192 | 0 | 132.471 | 22.065 | 21.093 |
| 0 | Original SQL | 168 | 24 | 329.819 | 27.719 | 31.794 |
| 0 | Prepared SQL | 180 | 12 | 292.892 | 29.492 | 35.832 |
| 1 | Off | n/a | n/a | n/a | 33.890 | 18.424 |
| 1 | Non-durable group4 | 192 | 0 | 106.729 | 24.080 | 21.909 |
| 1 | Original SQL | 176 | 16 | 307.297 | 27.197 | 31.749 |
| 1 | Prepared SQL | 186 | 6 | 280.892 | 27.467 | 29.555 |
| 2 | Off | n/a | n/a | n/a | 33.044 | 20.400 |
| 2 | Non-durable group4 | 192 | 0 | 118.219 | 26.802 | 19.852 |
| 2 | Original SQL | 178 | 14 | 289.943 | 27.883 | 33.777 |
| 2 | Prepared SQL | 177 | 15 | 310.260 | 26.816 | 36.238 |

Total completion rises from 522/576 (90.63%) to 543/576 (94.27%), with the
third trial losing one completion and worsening age. Prepared arms perform
27,150 original admission/readback/discard records. All arms have zero recorded
errors and entry expiries; losses are queue drops. All prepared read-p99/off
<= 1.10 screens pass; all age screens fail. Write p99 worsens against original
SQL in two trials and exceeds off in all three. No population tail or general
serving protection guarantee follows from these finite cells.

Prepared discard phase totals are 194/212/202ms versus 244/269/231ms, while
entry totals rise to 115/103/116ms versus 65/88/65ms. Counts differ by arm and
phases overlap other writers, so do not add isolated percentage speedups or use
aggregate totals as matched per-record cost estimates. End-to-end improvement
is mixed despite the isolated storage success.

## Audit and next lead

The full non-race load test passed accounting in 20.90s. Independent parsing
verified all twelve unique trial/mode/PreparedWrites cells, embedded hashes,
request/write counts, group-weighted conservation, 50 validations per completion
and exact durable operation counts.
Artifact SHA-256:
`2ff78c252ed88b7fb7f22cfad4309bf4a8fc3280f5de94ab8498a9691f92423f`.

Source inspection identifies another bounded repeated operation: each of up to
four frontiers independently fetches its event batch during service validation,
even when their event sets overlap under the same held guard. Next test a
per-guard union read while retaining every journal, query, baseline, feature,
identity and publication check. Do not truncate candidates or reuse query-specific
features across differing requests. Negative controls need different tenants,
times, queries, overlapping/disjoint event sets and missing/ambiguous returns.
Measure read counts and loaded results before claiming benefit.

Unique durable service identity, feedback/history authority, warm loaded learning
and realistic answer quality remain open alongside performance. All seven
research directions remain open. No production deployment, publication or push.
