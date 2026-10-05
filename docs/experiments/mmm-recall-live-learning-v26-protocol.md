# Eight-worker Recall plus durable live learning v26: frozen protocol

Date: 2026-10-01. This is a research-only Goal 6 integration test. The v25
eight-worker latency pass had no live labels; the earlier durable freshness
pass used a small serial frontier. Do not infer their conjunction without
running both paths on the same isolated store and writer workload.

## Workload and guards

Use the v25 200-live-event full Recall fixture, eight workers, 192 requests
offered nominally every 8 ms, `RecallK=200`, `PackK=10`, and 256 concurrent
future-only writes. Both arms use the same single guarded SQLite WAL/FULL
journal and durable publication lineage. Rotate `off` (no frontier consumer)
and `on` (bounded queue-64 consumer) over three independent store trials.
The `on` consumer chooses the predeclared `seed` event from every third
committed frontier and submits exactly 64 synthetic usefulness labels.
It must validate the committed journal/event/feature binding before durable
admission, revalidate before feedback, and observe successful live worker
publication. The label is a lifecycle fixture, not a truth or quality claim.

Measure full-request offer/call/queue p99, actual offer-gap p50/p99,
frontier-to-published and offered-feedback-to-published ages, admission and
feedback counts, drops, journal rows after close/reopen, and durable ledger
ordering/replay. Require exactly 192 completed offers, 200 distinct live
nominations each, no future packed event, 256 overlapping future-only writes,
192 acknowledged durable journals, zero tap drops, 64 successful live labels
with zero failed/pending work, and 128 ordered admit/feedback ledger rows
per `on` trial. After completion, one visible mutation must make an old
bound admission fail without executing its callback.

The `on` arm passes this **finite Goal 6 integration component** only if
its loaded offer p99 is <100 ms in each trial and pooled, pooled serving p99
is no more than 1.10 times matched `off`, and both age measures have pooled
p99 <250 ms. A failed invariant invalidates the performance result. Quiet
load, burst/rate sweeps, cross-epoch state transfer, crash/power-loss
recovery, real outcomes, and production/OpenClaw latency are outside this
test. Preserve a negative result; do not tune the queue or feedback cadence
after looking at the trials.
