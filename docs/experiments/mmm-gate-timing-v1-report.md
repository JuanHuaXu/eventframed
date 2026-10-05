# Direct publication-gate timing: attribution, not rescue

## Status

Goal 6 remains OPEN. The measurement locates foreground contention in the
research wrapper's validity-through-admission barrier. It does not demonstrate
a production performance fix or durable learning under load.

Protocol: [frozen contract](mmm-gate-timing-v1-contract.md).
Evidence: [raw captured run](mmm-gate-timing-v1.jsonl),
[summary](mmm-gate-timing-v1-summary.json), and
[independent invocation replay](mmm-gate-timing-v1-summary-replay.json).
The latter is a replay of the same analysis, not an independent experiment.

Raw SHA-256:
`cc5fec406b90ae2bb3080f7f740fb19bffbcef769b2c7d43615f0467e8fcb92d`.
All 812 embedded source hashes match both captured text and current files.
Recomputed summary is byte-identical. No experiment was rerun to obtain it.

## Method and accounting

Three rotated trials compare idle-wrapper and active-background arms, each
with and without bounded telemetry. Each cell offers 192 reads at 5 ms and
96 writes at 10 ms, with four native writer slots, queue capacity 64 and
groups of four. The idle control retains the wrapper: this is not the earlier
bare-native control. Percentiles use scheduled completion minus offered due
time, including queuing, rather than service time alone.

All 12 cells completed: 2,304 reads and 1,152 writes. All six active cells
accepted 192 observations each, with zero dropped or expired observations.
Each active cell reports 9,600 admissions and 9,600 discards; these are cold
observations, not successful model updates. Six measured cells recorded 729
spans without recorder overflow: 576 ingestion, 147 as-of, and six close spans.
Close spans belong to teardown and are excluded from foreground attribution.

## Scheduled latency

All values below are milliseconds; each entry lists trials 0, 1 and 2.

| Arm | Read p99 | Write p99 | Observation age p95 |
| --- | --- | --- | --- |
| Idle, unmeasured | 622.56 / 536.55 / 587.48 | 156.31 / 125.74 / 122.69 | N/A |
| Idle, measured | 384.50 / 511.61 / 549.28 | 73.79 / 107.22 / 113.55 | N/A |
| Active, unmeasured | 28.10 / 29.29 / 28.27 | 229.26 / 218.55 / 210.49 | 94.86 / 116.00 / 94.02 |
| Active, measured | 24.28 / 30.12 / 27.72 | 204.69 / 172.02 / 151.07 | 72.75 / 95.00 / 101.99 |

Active-background write tails exceed the matched idle tails in every trial,
both with and without instrumentation. Read tails improve, but this does not
offset the write regression or complete the joint serving requirement.
Measured arms often run faster than their unmeasured counterparts. Three
rotated trials cannot separate instrumentation cost from order, cache and
scheduling variation; this is not a telemetry speedup or zero-overhead claim.

## Where writes wait

Measured totals per cell, in milliseconds:

| Trial | Idle ingestion wait | Active ingestion wait | Active as-of held | Active durable admission |
| --- | --- | --- | --- | --- |
| 0 | 0.0120 | 359.40 | 358.53 | 266.76 |
| 1 | 0.0116 | 350.31 | 349.40 | 264.90 |
| 2 | 0.0150 | 361.70 | 360.63 | 268.28 |

Near-zero idle ingestion wait becomes 350-362 ms total with background work,
close to the as-of guard's 349-361 ms occupancy. Individual active ingestion
wait p99 is 17.14 / 19.44 / 22.38 ms. Durable admission accounts for roughly
74-76% of callback time. Callback, admission and held times overlap and must
not be added. Ingestion held time itself decreases from idle totals of
970-1,042 ms to active totals of 720-751 ms: the finding is added waiting,
not evidence that the native write operation becomes slower.

## Audit and limits

Normal constructors leave telemetry disabled. The bounded recorder stores
only durations and operation kinds, and records after releasing the gate.
Failed acquisition is separate from successful occupancy. Error and panic
cleanup retain the release path; optional vector capabilities are preserved.
The previous run passed full wrapper/service race tests and vet before the
19.02-second opt-in experiment. This checkpoint rechecked the timer/guard
implementation, source hashes and deterministic summary replay; it does not
claim those earlier suites were newly rerun here.

CONFIRMED: this fixture has guard contention and a write-tail regression.
NEEDS INVESTIGATION: how much a redesigned admission transaction can remove
while preserving validity, uncertain-commit recovery and complete accounting.
RECOMMENDATION ONLY: a staged, version-checked admission design. Releasing the
guard before durable admission is not an authorized or proven fix. Existing
small-group, SQL and exclusive-locking failures remain failures.

An alternative bottleneck remains native storage/scheduling contention, which
can coexist with the measured guard wait. A successful next design must reduce
scheduled write tails with all observations accounted for, preserve snapshot
and retry semantics under interleavings, and then survive outcome-bearing
learning tests. Moving work outside measured intervals or dropping observations
would not meet that requirement. No production, whitepaper or remote changes.
