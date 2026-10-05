# Actual shadow scheduler: completion failure

**FAIL in all three pairs.** Foreground latency non-harm passes, but completed
diagnostic jobs are1/64 per trial, below the predeclared52/64 requirement.
Each enabled arm records64 accepted,63 stale,1 completed,0 dropped/failed/cancelled.
All jobs reached a terminal state before closing the service. No read/write
errors; all16 writes overlapped readers in every arm.

| Trial | Off p99 ms | On p99 ms | On/off | Completed | Stale |
|---|---:|---:|---:|---:|---:|
|0|28.980|29.988|1.035|1/64|63/64|
|1|42.974|27.344|0.636|1/64|63/64|
|2|30.926|29.847|0.965|1/64|63/64|

This is actual Recall/Observe, the actual shadow queue, and a temporary
persistent LibraVDB backend, not a standalone worker queue. It uses synthetic
fixture text and a fixed labelled fitter workload, not labels derived from
the recalled events. Scalar shadow results cannot affect served answers.
Consequently completion is not semantic usefulness or learning validation.

The100ms age and ordinary snapshot-compatibility checks can both produce stale
status. Aggregate counters do not distinguish those causes, so it would be
incorrect to attribute all63 failures to fitting speed, queueing, or writes.
Future-dated writes were intentional; temporal reuse was disabled as declared.
The next test should distinguish invalidation causes or compare the existing
temporal-reuse guard under the same workload. Do not weaken freshness checks or
claim that faster fitting solved this failure.

Artifacts: `segment-transform-service-results.json` contains all384 read and96
write durations, six statuses, and source hashes. Verification:

```sh
node research/verify-segment-transform-service.mjs
```

The opt-in Go experiment passing indicates data collection completed; the
separate verifier correctly returns `passed:false` for the research screen.
Three small paired runs do not establish population p99. The p99 over64
requests is their maximum. Timer/context creation, retrieval and persistence
are included in service timings; no production process was touched.

Direction6 remains open. The result shifts the immediate next investigation
from tail arithmetic to scheduler freshness/completion behavior.
