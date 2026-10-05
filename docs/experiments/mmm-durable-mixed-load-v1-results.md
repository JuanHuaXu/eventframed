# Guarded durable feedback under concurrent future writes v1

The frozen [v1 contract](mmm-durable-mixed-load-v1-contract.md) was run from
test source SHA-256 `89b3fccd8f9c1ff40f6755cd82399555c5fd9d82e09df54856c2c3aaf340ca78`.
Both arms passed the source-binding, as-of, label, ledger-order and same-epoch
replay checks in the ordinary and race-instrumented runs. Each arm had three
trials of 64 Recall/feedback calls; the writer arm completed 768 future-dated
writes, of which 762 began during an active learning operation. No future
event entered a measured frontier.

| Run | Arm | Recall p50 | Recall p95 | Recall p99 | Recall max | Feedback p99 |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| Ordinary | Quiet | 5.48 ms | 6.38 ms | 7.35 ms | 7.59 ms | 1.00 ms |
| Ordinary | Future writer | 31.12 ms | 45.48 ms | 53.55 ms | 56.33 ms | 11.86 ms |
| `-race` | Quiet | 6.17 ms | 7.48 ms | 8.52 ms | 8.63 ms | 2.18 ms |
| `-race` | Future writer | 143.59 ms | 380.97 ms | 402.68 ms | 413.23 ms | 121.46 ms |

The ordinary writer-arm run passed the finite p99 < 100 ms screen. The
race-instrumented writer-arm run **failed that same gate**, making the v1
`go test -race` exit nonzero. No race detector warning or state/identity
assertion failure appeared. The gate was inadvertently applied to a timing
mode with high instrumentation cost; this does not erase the measured
slowdown, nor does it establish an uninstrumented population p99. V2 will
freeze the separation between correctness-under-race and an explicit
ordinary-build performance screen before modifying the test.

The ordinary writer p99 was 7.29x the quiet p99 (53.55/7.35), so contention
is material even in the passing finite screen. These synthetic future-only
writes do not validate visible backfills, deletes, mixed-source labels,
OpenClaw traffic, production throughput or general 100 ms service latency.
