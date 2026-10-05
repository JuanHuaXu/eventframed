# Profile partitioned derived-index construction

Diagnostic only: one3200 and one6400 arm with the no-sync-derived overlay and
unchanged offered-load parameters. CPU profiling starts after initial stores and
indices are built, and stops after request/compactor drain. Label goroutines read,
write and compact. Descendant library goroutines may inherit those labels.

Record runtime memory counters immediately before/after the measured interval.
Allocation profiles before/after support differential analysis; they are sampled
and can lag GC cycles. Capture the final allocation profile after an explicit GC
outside the timed interval. Do not treat profiled latency as validation. Preserve
fresh artifacts and authoritative reopen checks; no concurrent extra workloads.
