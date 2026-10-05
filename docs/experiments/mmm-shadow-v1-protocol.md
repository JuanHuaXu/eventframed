# Research shadow lifecycle/load pilot

Frozen 2026-09-12 before load measurement. Recommendation6 scaffolding, not
learner accuracy or complete production integration. Default disabled; only
programmatic service configuration enables the diagnostic callback. No CLI
switch, OpenClaw changes, private data or deployment.

The successful synthetic9-bit model is NOT applied to arbitrary real recall
features. Handoff contains snapshot, answer certainty, candidate count and at
most64 predictive scores by value. No IDs, text or mutable event references.
Callback output is one diagnostic scalar and has no path back into packing,
ranking, forecast law, or store updates through this API.

One worker, bounded channel, nonblocking enqueue/drop on full. Deadline includes
queue residence; snapshot checked before and after callback. Shutdown cancels
and waits, discards queued work, and prevents publishing after close. Callbacks
must honor cancellation; Go cannot safely force-kill an arbitrary goroutine.
Panics, nonfinite values and errors are rejected. No exception text is exported.
Status is version-stamped diagnostics, not an atomic current-state lease; the
store may advance after a read. No serving action may depend on these results.

Load test: two isolated memory-backed services, one disabled and one enabled;
four concurrent readers each perform128 recalls. A fifth goroutine inserts32
public fixture events. Begin with3 events. Same input sequences per service,
but schedules are nondeterministic; all durations preserved in test output as
summary quantiles. Alternate order over3 repetitions. Queue16, max age100ms.
Callback executes2048 passes of squared-score summation with cancellation checks,
as a bounded diagnostic workload, NOT the real learner or a neural network.

Report per-run p50/p95/p99 service Recall duration and errors, plus queue accepted,
dropped/completed/stale counters. Do not infer paired causal latency from three
nondeterministic schedule realizations. This does not include HTTP/network,
SQLite/libravdb, model fitting, actual agent generation, or production traffic.
No latency success threshold is declared for this diagnostic scaffold; actual
request p95/p99 budget and realistic learner workload remain required for item6.
