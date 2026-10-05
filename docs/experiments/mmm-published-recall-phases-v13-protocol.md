# Full Recall phase timing v13: frozen diagnostic

Date: 2026-10-02. V12 showed that eliminating empty-residual
lookups leaves the Goal 6 loaded response p99 near 500 ms. This
test-only probe restores the unchanged v10 service configuration,
corpus, publication/admission architecture, offer cadence and gates.
It adds per-Recall timestamps around adapter store calls, without
altering their return values or acknowledgements.

For every offered Recall, record admission wait, Search, predictive
graph read, selection/omitted-certificate reads, the wall interval
from first to last residual-candidate read, and journal owner wait,
durable append and publication. Record total occupied-call time and
the remaining uninstrumented service time. Residual workers overlap;
their summed call durations must not be subtracted from elapsed time.
Use wall interval instead. Report per-call p50/p99 and total
request-count coverage, together with existing correctness, writer
freshness and offer-to-done metrics. No zero-valued or absent phase
may silently count as measured.

Run two normal fresh trials. The unchanged v10 gates still apply and
an expected latency failure is preserved. Do not make a pass claim
from instrumented timing. The result must identify whether the
median critical path is mostly journal, Search/other reads, admission,
or uninstrumented service work, with uncertainty from instrumentation
and concurrency. A next architecture change requires a separate
frozen test that preserves durable acknowledgement and as-of order.
Production remains untouched.
