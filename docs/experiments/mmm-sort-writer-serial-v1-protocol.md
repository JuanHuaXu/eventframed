# Sort-key serial writer handoff v1: frozen private control

Create a private three-row sortable EventFrame collection and publish the
test-only READY marker at its exact latest LSN. Close the Store and SQLite
sidecar completely before opening a second Store on the same path. The second
Store writes one legacy unkeyed event, records its latest LSN, and closes.
Reopen the original gate and require: latest LSN equals the second Store's
recorded LSN and exceeds READY; the new event is durable; the stale marker
denies serving; and a full readiness scan refuses to republish an unkeyed
row. The reopened runtime snapshot must include all four events.

Run this finite control normally and under `-race`. Passing only supports
sequential ownership handoff and fail-closed marker behavior, not concurrent
multi-process writes, crash recovery, per-event restart performance, or
incremental marker renewal. Production remains untouched.
