# V45 complete mixed-load critical-path diagnosis

Prospective diagnosis, 2026-10-04. Run the unchanged sixteen-trial V44 owner-time
workload, with its complete offers, durability, native backend, epoch guards and
both eager arms. Collect CPU, mutex, blocking and sampled allocation profiles.
Do not interpret profiled timings as adoption results: profiling adds overhead.
Do not reduce work, drop visible writes or remove outcome processing to pass.

Competing causes: retrieval/scoring CPU; authority or native backend contention;
durable storage waiting; post-request trace serialization/fixture setup; offered
load exceeding sustainable service capacity. A whole-process CPU percentage
cannot by itself distinguish these. Separate cumulative caller stacks for
Recall, outcome and publication from the fixture/output envelope. Blocked time
is summed goroutine time, not CPU time or a single request's elapsed latency.

Reuse the existing independent full trace checker and its 52 corruptions. Freeze
exact source/protocol/runner bytes; retain the test binary, profiles, command
logs, raw bytes and host inventory. Report what the profiles cannot capture,
including C/native attribution and missing per-phase request wall time.
Any proposed patch must preserve full durability and as-of authority, identify
the first expensive transition, and have adjacent positive/negative controls.

This is a Goal 6 diagnosis, not a successful latency rescue, true production
profile, hardware-independent guarantee or completion of other whole goals.
Production/private corpora/whitepaper remain untouched; all seven goals open.
