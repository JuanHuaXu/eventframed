# Deadline-aware writer/admission scheduling v38: frozen protocol

Date: 2026-10-01. Research-only Goal 6 candidate after
[v37](mmm-recall-priority-gate-v37-results.md) reduced learner
age but failed both the <250 ms freshness and <=1.25 writer-p99
gates. This is a test-only alternative scheduler; the production
store guard and durable learner remain unchanged.

Keep the same external exclusive permit around each future-only
`Observe` and each guarded admission or feedback. An active task
finishes without preemption. When several tasks wait, grant the
permit to the earliest absolute *soft* deadline; break ties by
arrival sequence. Each selected admission and its feedback share
the frontier's tap-enqueue time plus200 ms as deadline. Each
future-write attempt has its offer time plus30 ms as deadline.
Expiration does not cancel or relabel work; it only changes
ordering. Every work item must eventually be attempted or
fail explicitly on context cancellation.

Run two new matched 4 ms full-Recall pairs with order deadline/
control then control/deadline. Retain eight workers, 192 full
200-event Recalls, 256 future-only writes, 64 selected bound
durable labels, queue64, selected channel32, reorder cap32,
admission channel16 and one guarded SQLite WAL/FULL journal.
Record actual offer gaps, Recall p99, frontier/feedback age p99,
writer offer-to-completion p99, service-call p99, total completion
time and writer starts during the offer window. Require exact
nomination, no-future/as-of, visible-mutation rejection, journal
reopen, durable replay, 64 labels and zero drops in every arm.

Use the **same frozen v37 gates**: candidate Recall p99 <100 ms
in each trial, pooled Recall p99 ratio <=1.10, pooled frontier
and feedback ages p99 <250 ms, 256 writes complete, pooled
writer offer p99 and summed total completion each <=1.25 times
control, and candidate writer starts during offers >=80% of
control. If any gate fails, preserve the failure. Unit checks
must show the earliest-deadline rule in both writer-first and
old-frontier-first cases. Do not tune deadlines on these runs.

A finite pass would still need burst, power-loss/cross-epoch,
real-agent and independent load confirmation. No production
change or whole Goal 6 completion follows from this screen.
