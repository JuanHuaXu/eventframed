# Pre-feedback phase attribution v32: frozen protocol

Date: 2026-10-01. Research-only Goal 6 diagnostic after
[v31](mmm-recall-cadence-v31-results.md) repeated a 4 ms
frontier-to-publication p99 above 400 ms. V31 located most age
between tap take and feedback offer but did not separate ordered
admission from waiting for the feedback worker.

Instrument only the v29 research consumer. For each of its 64
selected labels record three monotonic durations that must sum
exactly to the existing tap-take-to-feedback-offer duration:

1. tap take to start of guarded admission (journal lookup, selected
   handoff, reorder wait and setup);
2. guarded candidate admission, including durable admission and
   validation;
3. admission end to feedback offer (bounded channel send plus
   waiting behind earlier feedback work).

Run one fresh enabled-only full 200-event Recall fixture at 6 ms
and one at 4 ms, each with 192 offers, eight workers, 256
future-only writes, 64 labels, queue64, selected channel32,
reorder cap32, admission channel16 and one SQLite WAL/FULL journal.
Keep all exact nomination, no-future, as-of, mutation rejection,
durability/replay, phase-conservation and zero-drop checks. Report
actual offer gap, serving/freshness p99, p50/p99 of all three new
stages, and the per-label stages for the five oldest frontiers.
No threshold tuning, data suppression or learner behavior change.

The hypothesis is that at 4 ms the admission-end-to-feedback-offer
stage dominates the oldest labels because selected labels arrive
about every 12 ms while the single guarded feedback path takes
longer. It is falsified if order/setup or guarded admission instead
accounts for most of those labels' pre-feedback age. This is a
diagnostic, not a capacity pass or a production patch.
