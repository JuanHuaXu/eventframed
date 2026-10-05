# Durable phase diagnostic

Freeze before dispatch. Preserve the incremental overlay, local libravdb and
100ms scheduled-arrival workload. Restrict to visible writes with admission on:
ordinary/experimental packing,20ms/5ms read spacing, two repetitions (8arms,
256reads,128writes). Fresh retained databases and orderly reopen validation
remain, using the same seed/query/public content as DURABLE_OPENLOOP_PROTOCOL.

Wrap the actual store Search, PutBayesianJournal and Put entry points. Record
monotonic start offset relative to arrival, inclusive duration and error. Attach
per-request collectors through context with locking; copy rows immediately after
service return, before audit readback. Seed writes are not collected. Store
methods delegate exactly once; no phase bypass, extra retry or policy change.

The existing WaitNS records dispatch plus admission wait. Phase durations include
internal locks, serialization, I/O and index work: they are not pure disk time.
Compute total minus waiting minus these measured spans only as uninstrumented
remainder, not as pure CPU. Check spans do not overlap before summing. Concurrent
requests cannot be summed into an end-to-end wall duration.

These instrumented runs are diagnostic, not independent latency qualification.
Report successes and errors separately, including callbacks never entered. Do
not average away failed operations. Source hashes and raw spans are retained.
The next optimization must be selected from evidence at these boundaries, then
verified with fresh uninstrumented runs. No production changes.
