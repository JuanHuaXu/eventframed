# Incremental diversity service replay and fixed-arrival comparison

Freeze before runs. The new overlay copies the prior task-lexical overlay and
changes only Select's call from diversify to researchDiversifyIncremental.
Both original and experimental packing arms use the faster implementation.
The original helper remains as the reference for exact-output unit tests.

First run the packing and researchcalendar tests with the race detector under
this overlay. Then replay all five public datasets with the same local embedding
model and parameters as TASK_LEXICAL_SERVICE_PROTOCOL.md. Compare all108 outputs
against stored lexical-service-results.json: selected candidate/law/score,
support rank, before/after rank traces, calibration flags, and explanation JSON
must remain equal. Timings may differ. No new learning or holdout accuracy claim.

Then repeat TASK_LEXICAL_OPENLOOP_PROTOCOL.md unchanged: 200 candidates, four
permits, original/experimental packing, admission off/on, future/visible writes,
50+25/s and200+100/s schedules, two repetitions,32reads+16writes per arm.
The reused32-arm grid is consumed performance-design data, not a fresh semantic
confirmation. Keep100ms scheduled-arrival deadlines and strict no-error/no-stale
criteria. Do not move readbacks, drop slow calls, or relax freshness validation.

Preserve new artifacts and source hashes separately. Compare against the earlier
fixed-arrival run descriptively; these are separate executions, not simultaneous
paired measurements. The microbenchmark's7x figure does not predict the service
speedup. Full rescue of this grid requires all admitted arms to pass; otherwise
report remaining failures. Passing still does not qualify durable production
storage, long-run stability or general agent usefulness.
