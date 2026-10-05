# Conditional insert v66 results

## Verdict

FAIL. Conditional INSERT SELECT is slower for every paired fresh and retry
cell. Fresh200-record admission regresses37.76-40.85%; retries regress20.86-22.87%
at200 records. Both frozen performance requirements fail. Do not integrate this
strategy into serving or change the thresholds. Correctness passing is not a
performance rescue.

## Evidence

[Protocol](mmm-conditional-insert-v66-protocol.md),
[raw artifact](mmm-conditional-insert-v66.jsonl), SHA256:
`6e623891623bbbdab284d65758dd4c4b6b611da920bb07905edce5bcf3fee74f`.
All35 captured source hashes match embedded and local files. Twenty-four cells,
96000 fresh originals/exact retries and96000 verified terminals,768 synthetic
setup labels. All paired original/training hashes and final database sizes
match. Reopen/lifecycle/terminal checks pass. Go1.27.1, Apple M4,darwin/arm64,
GOMAXPROCS10. No competing task-started performance tests.

Full ledger race suite PASS3.659s and memory race suite PASS13.922s. Additional
conditional owner uncertain-acknowledgment race test PASS1.369s: both before-
commit and after-commit errors stop the owner, and replay recovers the correct
fresh/retry result. Vet PASS. Ledger tests cover receipt/sequence parity,
concurrent and same-batch retries, late-conflict rollback, feedback ordering,
source-identity uniqueness, cancellation and panic. This is not hardware
power-loss validation. Experiment PASS8.379s means integrity, not speed.

## Timings

Ranges of paired trial means,ms. All cells have32 cycles. Baseline is resolved
source with prepared lookup-then-insert; candidate changes only batch strategy.

| Batch | State | Fresh baseline | Fresh conditional | Retry baseline | Retry conditional |
| --- | --- | --- | --- | --- | --- |
| 50 | cold | 1.011-1.219 | 1.347-1.390 | 0.814-0.912 | 0.973-1.011 |
| 50 | trained | 0.977-0.988 | 1.287-1.316 | 0.811-0.818 | 0.989-1.004 |
| 200 | cold | 3.528-3.584 | 4.917-4.972 | 3.158-3.186 | 3.856-3.870 |
| 200 | trained | 3.453-3.527 | 4.839-4.864 | 3.074-3.121 | 3.772-3.791 |

At50 records fresh regression is14.06-34.64%, retry regression10.92-22.93%.
Unlabeled cleanup still follows the existing feedback path; its timings are
retained in raw data. No observation has been converted to a label or replaced
with a precomputed forecast to obtain these comparisons.

## Why fewer calls did not help

A post-result EXPLAIN diagnostic in conditional_plan_test.go compares the
actual SQLite VM plans with the service-identity index enabled. Conditional
INSERT SELECT introduces OpenEphemeral, a coroutine, an extra Insert and extra
read/row traversal opcodes absent from plain VALUES insertion. The test passes
and logs both complete opcode-count maps. This confirms additional internal
work despite removing one Go/SQL statement boundary. It is an explanatory lead,
not a timed attribution of the entire regression to one opcode. Diagnostic code
was added after the measured artifact and is not one of its captured sources.

The candidate remains an explicit unused research constructor/API. It preserves
FULL WAL, transaction rollback, actual worker forecasts, source guards and
commit-before-ack. Default constructors still select the previous strategy.

## Next lead

Stop pursuing per-row self-referential INSERT SELECT for this workload. A
bounded multi-row VALUES insertion with one transactional preflight query is a
different hypothesis: amortize statement boundaries over an all-admission batch
without self-reading inside each insertion. It must preserve ordered receipts,
same-batch retries, source uniqueness and rollback, and fall back for mixed
admission/feedback ordering rather than silently reorder it. Test these contracts
before any timing or default change. All research directions remain open.
