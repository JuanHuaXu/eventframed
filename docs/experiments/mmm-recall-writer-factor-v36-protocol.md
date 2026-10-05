# Future-writer factor for guarded admission v36: frozen protocol

Date: 2026-10-01. Research-only Goal 6 causal diagnostic after
[v35](mmm-recall-guard-entry-v35-results.md) found a 14–15 ms
median before entering the guarded source-validation callback.
The shared `researchpublicationstore.Store` writer gate may be
contended by the fixture's 256 future-only `Observe` calls, but
v35 did not separate that from snapshot/lineage compatibility.

At each nominal offer gap 6 ms and 4 ms, run one fresh enabled
learning fixture with 256 concurrent future-only writes and one
fresh enabled fixture without that writer. Use order writer-on,
writer-off at 6 ms; writer-off, writer-on at 4 ms. Everything
else is identical: 192 full 200-event Recalls, eight workers,
64 selected bound durable labels, queue64, selected channel32,
reorder cap32, admission channel16, single guarded SQLite
WAL/FULL journal, exact nomination/no-future/as-of, journal
reopen, durable replay and visible-mutation rejection after the
offers. The writer-on arm must record exactly 256 writes with
offer overlap; the writer-off arm must record zero writes.
Both require zero tap drops and all 64 labels. Report actual
offer gaps, serving and frontier-age p99, guard-entry and
candidate-validation p50/p99, plus the five oldest labels'
guard-entry times. Keep the v35 test-only context marker and
all exact phase-conservation checks.

Prediction: if future-writer contention dominates, removing
those writes substantially lowers guard-entry time and may
restore <250 ms frontier freshness at 4 ms. If not, this
explanation is rejected. A writer-off pass cannot establish
the real writer-on Goal 6 requirement or authorize disabling
writes. No production code is changed.
