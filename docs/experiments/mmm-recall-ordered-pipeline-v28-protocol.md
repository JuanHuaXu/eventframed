# Bounded ordered learning pipeline v28: frozen protocol

Date: 2026-10-01. Research-only Goal 6 candidate following v26/v27.
V27 measured tap-wait p99 418 ms at one selected label per three offers,
and a race-stressed run confirmed that committed frontiers can arrive out
of `AsOf` order. A serial consumer's selected-label admission and feedback
medians sum to about 30 ms against a nominal 24 ms selected-arrival interval.

Use the same eight-worker, 200-live-event full Recall fixture, 192 offers
at a measured ~8 ms cadence, single guarded SQLite WAL/FULL journal,
queue-64 frontier tap, 256 concurrent future-only writes, and exactly 64
synthetic lifecycle labels. Compare learning off with learning on over three
rotated trials. Select offered request indices 0,3,...,189, not every third
completion. A research-only coordinator identifies request index from the
committed journal session ID and holds at most 32 selected out-of-order
frontiers; it admits selected predictions in increasing offered-index/
`AsOf` order. A second goroutine performs guarded feedback in that same
order. The admission channel is bounded at 16. No label may be used before
its own committed prediction, and a later forecast may omit earlier
unprocessed feedback but must never use future feedback. Never backdate a
prediction to suppress an error. Missing selected frontiers, exceeded
reorder capacity, or tap drops fail closed in this finite fixture.

Require all v26 serving, as-of, durable journal, visible-mutation, 64-label,
zero-drop and replay checks. The durable ledger may interleave admissions
and feedback, but each ID must have exactly one admitted prediction before
exactly one terminal feedback; admissions and terminals must each be in
ascending ID order, with bound source/event and declared synthetic labels.
Record actual offer gaps, offer p99, tap-wait and frontier-to-published ages,
and stage timings. A finite component pass requires on-arm offer p99 <100 ms
in every trial and pooled, pooled p99 no more than 1.10 times matched off,
and pooled frontier and feedback ages p99 <250 ms. Run a focused `-race`
trial for as-of/lifecycle correctness; race-build timings are excluded from
the ordinary performance gate.

This fixture has a known finite issuance set, so ordered completion can be
checked without a general production watermark. A pass would **not** prove
handling of never-arriving frontiers, arbitrary event-time lateness,
cross-epoch transfer, power-loss recovery, real outcomes, or OpenClaw
serving. Preserve a failure without weakening the v26 freshness threshold.
