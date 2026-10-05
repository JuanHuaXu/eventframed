# Bounded admission-priority writer scheduling v37: frozen protocol

Date: 2026-10-01. Research-only Goal 6 intervention after
[v36](mmm-recall-writer-factor-v36-results.md) isolated the
effect of 256 future-only writes on guard entry. This is an
external test-only scheduler, not a production store change.

Use a single mutually exclusive permit around each future-only
`Observe` call and each guarded research admission or feedback
call. When a research operation is waiting, a new writer waits;
an already-started writer finishes. Research admission and
feedback each retain their existing order and as-of/durable
guards. The scheduler is context-cancellable and holds no permit
across Recall serving, tap drainage, or worker publication.

Compare uncoordinated control and admission-priority candidate
over two fresh matched pairs at nominal 4 ms offers, with order
control/candidate then candidate/control. Every arm keeps eight
workers, 192 full 200-event Recalls, 256 future-only writes,
64 selected bound durable labels, queue64, selected channel32,
reorder cap32, admission channel16, and one guarded SQLite
WAL/FULL journal. Record actual offer gaps; Recall offer p99;
frontier and feedback ages p99; tap drops; writer calls, attempts
during offer window and calls started during offer window;
writer offer-to-completion p99, service-call p99, and total wall
time from first writer attempt to completion of the 256th write.

All arms must pass exact nomination, no-future, as-of,
visible-mutation rejection, journal reopen, durable replay,
64 labels and zero drops. A candidate finite component passes
only if every candidate trial has Recall offer p99 <100 ms,
pooled candidate/control Recall p99 <=1.10, pooled frontier
and feedback ages p99 <250 ms, all 256 writes finish, pooled
writer offer-to-completion p99 and total writer completion
time are each <=1.25 times matched control, and the number of
writer calls actually started during offers is at least 80%
of control. This last gate prevents deferring the writer until
after the query window. Do not retune those margins after the run.

The intervention fails if it improves learning only by starving
writers, leaks future outcomes, loses selected work, or breaks
durability/as-of rules. A passing finite screen would not prove
arbitrary bursts, cross-epoch recovery, power-loss behavior,
real-agent outcomes or whole Goal 6 completion. Production
and publication remain untouched.
