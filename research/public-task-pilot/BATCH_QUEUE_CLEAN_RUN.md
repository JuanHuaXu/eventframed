# Scheduling correction before rerun

The first batch-queue-results.json run overlapped a race-test command near its
final arms. Retain it as diagnostic-only: do not cite its latency as isolated
performance evidence. Both process handles have completed (load38524, tests81939).
The combined race suite passed. No changes to the queue, runner or store are
being made for the rerun.

Repeat the exact32-arm protocol to batch-queue-clean-results.json with no agent-
started tests, benchmarks or builds running concurrently. Build startup precedes
timed requests. Other host activity is not controlled. This is an uncontaminated
rerun relative to our race tests, not a claim of a dedicated host. Do not rename
the old result, suppress failed arms or treat this as independent semantic data.

Project-local workflow lesson: await completion of every test process before
starting a measured load run; while it runs, restrict work to light inspection
and documentation. Record overlapping work and rerun to a new artifact if this
boundary is accidentally crossed. No global instruction changes are needed.
