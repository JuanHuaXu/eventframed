# Queued exact-snapshot guard v21: primitive verified, load not yet tested

## Reasoning and scope

V19 rejected1,152 observations busy; v20 admitted all read-only observations but
none under writes, regardless of journal-prefetch placement. Immediate try-lock
offers no queued admission opportunity. The root hypothesis for this prototype
is entry starvation under concurrent writes, not corrupted snapshots or bad
service bindings. Successful controlled v18 callbacks and v20 read-only cells
are the adjacent-path controls.

The unconfigured research adapter now uses a one-token `x/sync/semaphore.Weighted`
for the SAME shared writer ownership boundary. Existing mutations/Close retain
unconditional serialized acquisition; the existing immediate guard retains
nonblocking acquisition. `WithResearchSnapshotWait` adds cancellable queued
acquisition with a REQUIRED caller deadline. This uses the already-pinned
dependency's waiter/cancellation implementation, not polling or detached waiter
goroutines. No daemon constructor enables the adapter or the new method.

Once admitted, both guards use the same under-lock exact snapshot and quarantine
check. Waiting does not confer validity: a policy change or even compatible
future-only motion can still cause exact-snapshot rejection. An as-of-compatible
guard has NOT been added. The deadline bounds acquisition, not a callback that
ignores cancellation; callback owners must bound work and must not mutate or
close the adapter from inside the guard.

## Verification

Targeted queued/immediate guard and persistent-boundary tests passed three race
repetitions. Full publication-store, service and research-memory race suites
passed once; vet passed for all three packages. Tests cover missing deadline,
timeout while owned, entry after release, exclusive ownership during callback,
stale rejection, callback errors and panic release. Existing mutation/rollback/
capability tests remain part of the full suites.

This supports lock integrity in tested cases, NOT loaded availability, strict
worst-case latency, fairness under arbitrary workloads or persistent-learning
success. The shared semaphore changes research-adapter lock overhead, so old
timings must remain attached to their embedded original sources.

## Next evidence required

Run same-workload immediate/queued admission comparisons with a predeclared
wait budget, distinguishing timeouts, stale snapshots and accepted work. Do not
count a busy-to-stale transition as success. Only then consider under-lock
as-of compatibility with policy/backfill negative controls. Full-feedback
authority, validation batching and end-to-end request/observation-age budgets
remain required; v20's slow read-only drain must not be hidden by better entry.
