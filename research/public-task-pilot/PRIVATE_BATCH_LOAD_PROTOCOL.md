# Bounded collector load rescue

Use PRIVATE_GRAPH_LOAD_PROTOCOL unchanged for workload, resources, deadlines and
serving/durability gates. Candidate combines norm reuse with atomic insertion
batches: at most four, 10ms collection window, 64 outstanding calls including
the active batch. Earliest member deadline governs the whole batch; cancellation
of a participating member cancels the batch. Expired jobs are removed before
preparation. No retries, no acknowledgements at enqueue, no weakened quarantine.

Two complete repetitions, retaining raw outputs even on failure. Collector Close
drains and joins before writer/storage Close. BatchID and BatchSize identify
shared preparation/persistence timings: do not sum repeated timings across batch
members. Queue wait remains per-request, total NS remains scheduled-arrival
latency. A returned success after deadline is counted as late, not hidden.

This combined candidate cannot isolate batching's benefit from norm reuse.
The previous singleton load and norm screen remain failed. No thresholds or
seeds may be changed after inspecting this run.
