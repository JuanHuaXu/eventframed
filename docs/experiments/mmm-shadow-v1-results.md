# Shadow handoff lifecycle pilot

2026-09-12. [Protocol](mmm-shadow-v1-protocol.md),
[final load output](mmm-shadow-v1-load-final.txt),
[service regression rerun](mmm-shadow-v1-service-regression.txt).

## Implemented boundary

The actual Service.Recall completion path now has an opt-in, default-disabled
research diagnostic handoff. It copies at most64 scores plus count, certainty
and snapshot by value. It carries no text, event/session/tenant IDs or mutable
event references. Served packets never consume callback results.

One worker, capacity1..256, deadline including queue residence, nonblocking drop,
pre/post snapshot checks, cancellation, bounded result state and shutdown drain.
The callback is a trusted in-process extension and must honor cancellation.
The API is not a sandbox preventing arbitrary callback code from accessing
ambient process resources. No HTTP endpoint or CLI flag was added.

This is NOT an implemented real-feature version of the synthetic9-bit learner.
That feature/likelihood/feedback mapping remains unvalidated. The pilot callback
does bounded numeric diagnostic work, not prediction training. No production
configuration, OpenClaw installation or persisted memory was changed.

## Verification

Targeted race tests passed for packet equality versus disabled control, immutable
handoff, queue pressure, cancellation/drain, in-flight snapshot invalidation,
timeouts, panic/error/nonfinite rejection, and result clearing after shutdown.
Vet passed. The load test also passed under the race detector (timings from that
instrumented run are not used as performance evidence).

Audit found that shutdown initially cleared the valid-result flag but left its
old numeric value visible. The value now clears too; a regression checks that.
The initial load output is preserved as `mmm-shadow-v1-load.txt`; final output
was rerun after this diagnostic-only repair.

## Diagnostic service latency

Three enabled and three disabled isolated memory-store runs,512 recalls each,
four readers plus an initial32-event write burst. Final uninstrumented p99:

| Repetition | Disabled microseconds | Enabled microseconds |
| --- | ---: | ---: |
| 0 | 1008.541 | 633.292 |
| 1 | 649.167 | 666.250 |
| 2 | 720.916 | 650.041 |

Zero recall errors. Each enabled run accepted/completed512 jobs with no drops
or stale results. This small workload did NOT saturate the queue. Explicit
pressure and in-flight-write tests, not these counters, establish those branches.
The write burst may finish before most recalls; sustained write contention is
not established. Scheduling/warm-up noise prevents interpreting lower timings
as a speedup. No production-latency gate is passed by this scaffold.

Missing: real learner workload and validated feature mapping, actual feedback,
HTTP/network, SQLite/libravdb, sustained writes, longer backlogs and request
p95/p99 under representative traffic. These remain required for recommendation6.

## Unresolved regression observation

The first full service suite run failed the existing
TestEventFrameDeltaReranksAfterRetrievalContract while ResearchShadow was disabled.
Observed baseline score0.955030827 exceeded learned Bayesian probability
0.954545455, producing a negative rather than the test's expected positive
rank delta. The test uses time.Now inside event construction; timestamp-dependent
embedding variation is a plausible explanation, not a proven root cause.

Three isolated repetitions passed, then the full service rerun passed. This is
classified NEEDS INVESTIGATION/intermittent, not fixed, and not evidence that
the shadow hook caused it. No unrelated ranking behavior or assertion was changed.
Overall production regression readiness must not be called clean from one rerun.
