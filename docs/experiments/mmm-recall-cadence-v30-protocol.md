# Full-Recall cadence envelope v30: frozen protocol

Date: 2026-10-01. Research-only Goal 6 capacity screen after
[v29](mmm-recall-decoupled-drain-v29-results.md). V29 passed a finite
eight-worker 8 ms (~125 offers/s) fixture, but its burst headroom is
unknown. The candidate changes only the test offer cadence; the
decoupled bounded tap drainer, guarded admission, durable feedback,
and production code remain identical.

Run one matched learning-off/on pair at each nominal offer interval:
8 ms control, 6 ms (~167 offers/s), and 4 ms (~250 offers/s). Use
arm order off/on at 8 ms, on/off at 6 ms, and off/on at 4 ms. Keep
192 full 200-event Recalls, eight workers, 256 future-only writes,
tap capacity 64, selected-frontier channel capacity 32, reorder map
limit 32, admission channel capacity 16, one SQLite WAL/FULL journal,
and 64 synthetic bound labels. Each pair gets a fresh isolated store.
Measure actual median/p99 offer gaps, per-arm offer p99, on-arm
frontier-to-publication p99, feedback-to-publication p99, drop count,
and lifecycle completion. A `t.Fatalf` from a fixture is a failure
for that arm, not a reason to silently skip its rate.

The fixed finite screen passes a rate only if the offered gaps have
median within 25% of the target, both arms finish all 192 Recalls,
the enabled arm consumes 192 unique frontiers, publishes 64 labels,
and has zero drops; all exact nomination, no-future, as-of,
visible-mutation, journal-reopen and durable-replay checks pass;
on-arm offer p99 <100 ms, on/off offer-p99 ratio <=1.10, and both
on-arm frontier/feedback ages p99 <250 ms. A rate failing any gate
is outside the demonstrated finite envelope; do not loosen gates
or retune on this screen. The 8 ms control should be consistent with
v29 but is an independent run, not a recomputed v29 cohort.

This is one screening pair per rate, not a robust capacity claim.
A passing higher rate needs fresh repeated confirmation with a
frozen burst shape and fixed machine load. A failing rate needs
phase attribution before a performance patch. Missing selected
frontiers must fail closed rather than be silently relabeled or
replayed from future state. No user text or production data is used.
An adjacent negative control closes an empty tap before consumption
and requires a terminal error with zero admitted labels.
