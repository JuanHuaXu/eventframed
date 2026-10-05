# Live-learning age and order diagnostic v27: frozen protocol

Date: 2026-10-01. This is a research-only Goal 6 diagnostic following the
failed v26 freshness gate. It does not change the v26 workload, selection
rule, learner, durable journal, feedback schedule, or as-of guard. The v26
test source is instrumented only to expose stage durations and the last
successfully admitted label's availability time on a failure.

Run three enabled eight-worker trials with 200-event full Recall,
192 requests at nominal 8 ms intervals, 256 concurrent future-only writes,
queue-64 tap, and one bound durable label from every third committed
frontier. Keep the v26 exact nomination, no-future, no-drop, journal-reopen,
64-label publication, 128-row replay, and visible-mutation rejection checks.
For each selected label measure:

1. `tap_wait`: frontier enqueue to consumer take;
2. `pre_feedback`: consumer take through validated admission to feedback offer;
3. `feedback_guard`: feedback offer through guarded feedback return;
4. `publication_wait`: guarded feedback return to observed worker publication.

Record p50/p99 and conservation against frontier-to-published age. If
`tap_wait` dominates, investigate queue capacity and consumer concurrency;
if `pre_feedback` dominates, investigate journal validation/admission; if
`publication_wait` dominates, investigate worker fitting/notification.
These are phase-attribution hypotheses, not permission to weaken the 250 ms
freshness or <100 ms serving gates. A focused race-instrumented replay may
fail, but it must report current frontier `AsOf` and the prior label's
availability so a backdating explanation can be checked directly.
No phase timing from a race build is used as ordinary performance evidence.
