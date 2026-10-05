# Full Recall worker-capacity diagnostic v24: frozen protocol

Date: 2026-10-01. Research-only Goal 6 diagnostic; no production changes.
The v23 no-wait group arm had 1.10 s queue p99 under 8 ms offered arrivals,
but p99 call duration cannot establish average capacity. This follow-up
uses the same full Recall 200-event fixture, exact nomination/as-of/reopen
checks, four workers, 192 offers nominally 8 ms apart, and 256 concurrent
future-only writes. Compare single guarded SQLite, 8 ms group dwell, and
1 ns group dwell over three rotated trials. Reuse the existing v22/v23 trial
functions; do not change their persistence behavior or source.

For each arm/trial, record mean call duration, call p99, queue p99, mean
journal span, mean guard-wait span where available, mean SQLite insert span
where available, and a diagnostic first-vs-last-completed-quarter queue
median. `recallProfileTrial` records samples in completion order, so the
quarter comparison is **not** an exact offered-index time series.

The nominal arrival rate is 125 offers/s. Four workers require mean call
duration below 32 ms to sustain that rate before accounting for jitter or
other work. If all three loaded trials of an arm have mean call >32 ms and
queue growth, classify it as nominally capacity-limited on this fixture. If
mean call is below 32 ms but queue grows, reject that simple capacity
explanation and investigate burstiness, writer lock timing, or uneven
worker occupancy. Any source-time jitter or overlap with post-offer drains
limits the inference; this is not a throughput theorem. No latency success
is declared by this diagnostic. Preserve the v23 failure and <100 ms gate.
