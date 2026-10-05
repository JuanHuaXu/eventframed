# Notification load repeat v9

2026-10-01. On the **current** research wrapper, queue-64 notification passed
all six finite load cells again. This is a repeated workload, **not** a
byte-identical replication of v8's code. Four dependency hashes changed since
v8: `background.go` adds restored-worker ownership, `publication.go` adds a
committed-snapshot check, `capabilities.go` adds measured wrapping, and the
publication-store writer gate changed from `sync.Mutex` to a cancellable
one-token semaphore. The test file, workload shape and frozen gate are
unchanged. These source differences could affect contention and scheduling.

The [protocol](mmm-notification-repeat-v9-protocol.md) was frozen before the
run. All 36 arms completed: queue 16 or 64, six rotated trials, off/notified/
polling, 192 recalls and 96 future writes per arm, four readers, and 50 labels
per admitted frontier. The test process exited nonzero because five **polling**
cells failed their frozen gate; those failures are retained. Every notified
cell passed in this run, including queue 16, whose v8 trial-0 p99 failure is
not erased by the later result.

| Queue | Mode | Passes | Admissions across six trials | Age p95 range (ms) | Failure type |
|---:|---|---:|---|---:|---|
| 16 | Notified | 6/6 | 162-167 / 192 | 85.98-90.97 | None in v9; v8 had one p99 failure |
| 16 | Polling | 4/6 | 150-158 / 192 | 90.45-103.27 | Two admission misses |
| 64 | Notified | 6/6 | 192 / 192 | 160.19-204.18 | None |
| 64 | Polling | 3/6 | 192 / 192 | 220.02-272.09 | Three age misses |

The queue-64 notified serving-p99/off ratios ranged 0.823-0.950, below the
predeclared 1.10 ceiling. These ratios are workload measurements, not proof
that enabling notification speeds serving. All arms had zero reported errors,
zero worker failures and overlapping writes. Every admitted frontier completed
all 50 labels. The independent checker verified current/embedded source hashes,
36 distinct arms, all read/write samples, nearest-rank p95/p99, admission/drop
and label conservation, and each frozen per-cell gate. It found 19/24 enabled
passes overall (12 notified, 7 polling).

The [raw JSONL](mmm-notification-repeat-v9.jsonl) has SHA-256
`99e9e9b8933c1387f8025cbf15682b9bdd80bada69387692a1218cb2aa3dffc2`;
the [independent check](mmm-notification-repeat-v9-check.json) records every
cell and source hash. Reproduce with `node research/check-notification-repeat-v9.mjs
docs/experiments/mmm-notification-repeat-v9.jsonl NEW_OUTPUT.json`.

Decision: queue-64 notification remains a research candidate for goal 6, not
a production setting. Six trials on one synthetic arrival pattern do not show
population p99, mixed-mutation correctness, recovery of durable feedback,
model quality, or fresh agent-task improvement. The next test should hold the
current wrapper fixed while adding realistic mutation/restart/feedback
conditions, rather than treating this repeat as a completed integration.
