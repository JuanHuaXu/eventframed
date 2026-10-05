# Archive Lifecycle V31: Operation Timer Placement

V29 frozen barrier/ancestry controls PASS. V30 interruption/ownership, vet and
core race checks PASS, but sixteen concurrent archives FAIL: `read durable
migration marker: context deadline exceeded`. Preserve V30 sources and logs.

Confirmed upstream cause: V29 starts its500ms database-operation timer before
`commit` takes the owner mutex; queued archives enter persistence with expired
contexts. This differs from sealed V25's timer placed after owner acquisition.
Alternative explanations (native stale LSN, byte readback and history corruption)
are not supported by V30's deadline error; interrupted-commit controls remain
required. No production bug or measured latency benefit is inferred.

New research wrapper only: accepted captures wait for owner using cancellation-
detached context, then start the SAME500ms operation timer immediately after
owner acquisition, before ancestry/native/sidecar work. Queue time is unbounded
in this prototype and must remain included in offered end-to-end metrics; this
does NOT relax100/250ms adoption gates or establish overload safety. Cancellation
still returns no packet and waits for accepted capture terminalization.

Mechanically clone V29 commit with just timer placement and V30 tests with new
attachment. Freeze both generators, generated files, adapter, all runtime Go,
protocol and runner before the run. Rerun interruption, sixteen concurrent
handoffs, ownership, vet and three repeated scheduler/validity race checks. Keep
V29/V30 artifacts unchanged. Falsifier: queue-induced operation expiry persists,
current/archived law changes, a missing ack, failed reopen or weakened negative
control. No loaded latency, quality or general lifecycle completion claim.
