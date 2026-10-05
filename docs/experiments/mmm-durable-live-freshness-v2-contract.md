# Offered-label to live publication v2

Frozen 2026-10-01 after the [v1 endpoint correction](mmm-durable-live-freshness-v1-results.md)
and before changing the test. Retain v1's two arms, three persistent trials
per arm, 64 chronological guarded labels per trial, 256 maximum concurrent
future writes, as-of/source/ledger/replay assertions, and the same finite
thresholds: ordinary-build writer Recall p99 <100ms and live age p99 <250ms.

The sole measurement change is the origin and subscription point. Record a
monotonic timestamp and enqueue the target ID to the independent notification
observer **immediately before** entering the guarded feedback call. The
observer waits for that ID's absolute worker completion count, checks zero
failures, and records the first time it observes publication. If the worker
publishes before the guarded call returns, the observer may see it then.
Observer scheduling can still inflate measured age, so this is a conservative
offer-to-observed-publication bound, not an exact internal fitting duration.

The observer must witness all 64 published labels before Close. Keep the
post-return v1 measurement and its limitation documented separately. Run
ordinary-build performance mode with `EVENTFRAME_RESEARCH_PERF_GATE=1`, race
correctness without that flag, focused Durable observation tests, vet and
diff checks. No production path change or cross-epoch claim.
