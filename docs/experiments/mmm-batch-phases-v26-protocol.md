# Guard phase diagnostic v26

Frozen before execution. Repeat all nine v25 arms using the same protocol and
caps. Add monotonic per-attempt timing only: queue age at Take, journal prefetch
plus request construction, guard entry (waiting plus native snapshot check),
callback duration and total guard duration. Record whether the callback entered
and whether it was accepted. No change to validation, scheduling or authority.

Check one phase record per attempt, nonnegative timings and, for entered work,
total >= entry + callback. Report p95 and summed durations by mode, plus the
original admission/drop and request-latency measurements. Queue age sums are
overlapping waiting exposures, not CPU time or elapsed experiment duration.
Guard entry is not pure semaphore wait. Instrumentation may affect timings.

Hypotheses: callback work remains dominant; entry serialization dominates;
or prefetch outside the guard dominates. Choose the next design based on measured
phase contributions rather than merely increasing queue size/deadlines. This
is not a new success criterion or predictive learning experiment.
