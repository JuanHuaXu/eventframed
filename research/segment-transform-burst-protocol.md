# Isolated fitting burst protocol

Diagnostic, not production serving validation. Compare frozen direct and
research transform full fitting on the same 272-frame delayed/missing-label
fixture, cap64, hazard.01, generic mass.95. Queue exactly12 simultaneous jobs;
1/2/4 workers; three paired trials, alternating arm order. GOMAXPROCS fixed at4.
No rejection, coalescing, persistence or early cancellation. Drain every job.

Record monotonic queue wait, compute duration and submission-to-completion for
every job. Report completion fractions by100ms and500ms from burst submission,
including queue wait. Late results still finish for accounting; this is not the
production stale-result publication policy. Do not report pooled tiny-sample
p99 or claim these jobs represent full requests. Warm shared tables before all
trials equally. Each worker uses its own result; no shared benchmark sink.

Success for this diagnostic: forecast parity and no reduced100ms/500ms counts
in any paired cell, plus at least10% median-makespan improvement at each worker
count. Failed cells remain visible; these criteria do not complete direction6.
Actual serving interference, durable publication and usefulness remain untested.

Output is exclusive-create JSON with individual job timings and source hashes.
Run via EVENTFRAME_TRANSFORM_BURST_OUT and the opt-in Go test. Normal test runs
skip it. No network, production processes, personal data or deployment changes.
