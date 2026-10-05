# Bound worker v11: loaded service screen

Date: 2026-10-01. [Frozen protocol](mmm-bound-worker-load-v11-protocol.md).
Research-only; production unchanged. **Result: FAILED / incomplete.**

The four-worker client offered 192 full Recall requests at 1 ms cadence while
64 guarded labels were offered. In two fresh quiet-arm attempts, nearest-rank
offer-to-return p99 was 356.99 and 385.84 ms, already above the frozen 100 ms
writer-arm target. The second attempt separated call p99 (18.12 ms) from queue
p99 (373.04 ms); offered-label-to-applied p99 was 6.00 ms. The first quiet
trial had 9.06 ms label-age p99. Neither attempt reached the full three-pair
matrix: the immediately following writer arm rejected its first admission with
`research event continuity guard busy`. The service's source-continuity proof
uses fail-fast `TryAcquire` while an owned future writer may hold the gate.
No writer-arm latency or freshness percentile exists; do not impute one.

A separately labeled **probe-only diagnostic** ran the same 192-request,
four-worker, 1 ms offer stream on three fresh stores without a learner or
writer. Offer p99 was 309.55, 332.81, and 318.65 ms; call p99 was 13.86,
13.11, and 12.12 ms; queue p99 was 298.51, 321.81, and 307.73 ms. Thus
queue overload is substantially present in the underlying Recall/journal path
at this offered rate. The quiet learner adds some contention, but this
diagnostic does not assign a precise causal share. The writer failure is a
separate availability problem, not proof of wrong source acceptance.

Reproduce the expected failing screen and passing sibling diagnostic with:

```sh
EVENTFRAME_RUN_MOTION_LOAD_V11=1 go test ./internal/service -run '^TestResearchMotionBoundWorkerLoadedScreenV11$' -count=1 -v
EVENTFRAME_RUN_MOTION_LOAD_V11=1 go test ./internal/service -run '^TestResearchMotionProbeOnlyDiagnosticV11$' -count=1 -v
```

The failing performance screen is opt-in so ordinary correctness suites remain
green. The next research step is **not** to change v11's gate: combine source
continuity and as-of validation under one deadline-bound writer guard, then
measure a predeclared rate matrix with offered-time queue accounting. A simple
retry of the old fail-fast check would leave a check-to-commit race. This is
not production authority or a population p99 result. Goal 6 remains OPEN.
