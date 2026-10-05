# Research admission lifecycle component

The request-lease load experiment motivated a reusable, research-only gate in
`internal/researchadmission`. It uses the existing weighted semaphore dependency:
one permit per reader, all permits per mutation. No daemon call sites or frozen
experiment sources changed. The earlier load results are not measurements of
this new component.

## Invariant and evidence

Every successful acquisition has a deferred release. Callback errors propagate;
panics propagate after release. Requests canceled while queued do not execute
their callback. Cancellation after callback admission is cooperative: the gate
must retain its permit until the callback actually returns. Otherwise a writer
could overlap a canceled reader still using the store.

Six top-level tests cover both read/write queued cancellation, error/panic release,
parallel readers versus exclusive writer, invalid capacity/nil callback and
already-canceled requests, active cancellation, and a separate-gate negative
control. Channel barriers and testing/synctest establish queued states without
wall-clock scheduling sleeps. All passed twenty repetitions with the race
detector; go vet also passed:

```sh
go test -race ./internal/researchadmission -count=20 -timeout=60s
go vet ./internal/researchadmission
```

This is lifecycle evidence, not a latency benchmark, fairness proof, production
integration, or proof that every service mutation participates.

## Required integration boundary

One gate must cover the full shared version/mutation domain. A per-request gate
provides no coordination; per-tenant gates are unsafe when store version changes
cross tenant boundaries. The negative control demonstrates that two separate
gates permit overlapping read and write callbacks.

Participating reads must hold admission through snapshot capture, computation,
and journal commit. Ingestion, deletions, posterior/residual updates, graph and
policy publication, and any other snapshot-invalidating mutations must share
exclusive admission. Background workers count as writers. Other processes are
not coordinated by this in-process gate. Keep store freshness validation even
when a gate is present; it detects mutations outside the admission domain.

Callbacks must not recursively acquire the same gate, retain protected work in
unjoined goroutines, or assume cancellation can forcibly stop a callback. There
is no automatic tenant registry or mutation interception. Future-compatible
writes still wait unnecessarily under this coarse policy.

Next: use this component in a new, separately frozen service experiment, then
test open-loop load and durable mutations with waiting included. The original
runner-only success does not close direction 6. All seven goals remain open.
