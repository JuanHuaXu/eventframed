# Live durable feedback freshness under future writes v1

Frozen on 2026-10-01 before the read-only completion API and test edits.
This addresses a limitation of the [mixed-write v2 screen](mmm-durable-mixed-load-v2-results.md): log replay after Close does not prove that a live frozen snapshot incorporated a label before Close.

Expose `Durable.WaitProcessed(ctx, target)` and `Durable.Counts()` as read-only
forwarders to the existing Background completion interface. The wait is for
completed+failed count, not success or source authority; the caller must
check failures. Do not hold the durable mutex while waiting and do not change
prediction, feedback, replay, publication, or serving behavior. Test the
wrapper's completed, canceled, and closed-before-target cases.

Use fresh persistent LibraVDB, publication wrapper, frontier tap and durable
learner in three trials for each of two arms: quiet and future-writer. Each
trial issues 64 chronological guarded recalls, bound admissions and explicit
labels with `i%3 != 0`; query i is at `now+2i seconds`, label availability
at `now+(2i+1) seconds`. The writer arm starts before the first recall and
attempts up to 256 unique `now+1h` events at 1 ms ticks. Require overlap.
As-of frontiers must contain only the original seed.

After each durable feedback returns, record the monotonic submission time
and enqueue its ID to one independent notification observer. The observer
waits for successive absolute completion counts, checks the failure count,
and records elapsed submission-to-published-snapshot age before Close. This
age is an upper bound if observer scheduling is late. The observer must
complete all 64 labels before Close; afterward, an independent ledger audit
and same-epoch replay must agree on all labels, source bindings and times.

Measure empirical Recall p50/p95/p99, guarded feedback p99, and live
completion-age p50/p95/p99/max across 192 labels/calls per arm. The finite
ordinary-build screen is writer-arm Recall p99 <100 ms and live completion
age p99 <250 ms, with zero worker failures or unobserved labels. Keep race
instrumentation as a correctness run with separately reported timings; it
must not enforce the ordinary-build latency screen. The threshold is local
to this fixture, not a population tail, OpenClaw, backfill, restart or
cross-epoch guarantee. Run focused race tests and vet; keep production
serving untouched.
