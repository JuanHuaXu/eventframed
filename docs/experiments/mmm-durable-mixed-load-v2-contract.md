# Guarded durable mixed-write load v2

Frozen 2026-10-01 after the [v1 result](mmm-durable-mixed-load-v1-results.md)
and before test modification. Keep the same two arms, three trials per arm,
64 chronological admissions and labels per trial, 256 maximum future-dated
writes, as-of/source/ledger/replay assertions, and measured metrics. Do not
change the workload or 100 ms finite Recall p99 threshold.

Separate two execution modes explicitly:

- Correctness/race mode: `go test -race` runs all checks and reports timings,
  but does not apply the latency gate to instrumented durations. A race
  warning or correctness failure still fails the test.
- Performance mode: set `EVENTFRAME_RESEARCH_PERF_GATE=1` on a non-race
  `go test` run. This applies the unchanged writer-arm p99 < 100 ms gate.

Report both modes, including the v1 race-instrumented gate failure. The
environment switch is only a measurement-mode selector, not production
behavior or a change to the performance threshold. Race timings are retained
as diagnostics, not merged with ordinary-build timings. Run `go vet` and
`git diff --check` after the code change. No production path changes.
