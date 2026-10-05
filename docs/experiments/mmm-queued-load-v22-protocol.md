# Queued exact-guard load v22

Frozen before execution. Compare current semaphore-backed immediate validation
and queued validation against off, three rotated trials. Match v19:192 recalls,
four readers,96 future writes spaced2ms,50 visible candidates,64-slot handoff,
fresh persistent stores per arm. No ledger or feedback fitting in either enabled
arm: first establish that useful callbacks can enter.

Queued acquisition has a20ms context deadline per attempt. The deadline bounds
entry and its initial under-lock check, not the callback's entire duration;
callback I/O retains the one-minute enclosing fixture context. No retries.
Both enabled arms preserve exact snapshot/quarantine checks and full50-candidate
validation. No as-of-compatible relaxation is present.

Record timeouts separately from busy/stale rejections, queue drops and complete
accepted observations. Account attempts+drops=192 and
attempts=accepted+busy+stale+wait_expired; validated=50*accepted. Report request
p95/p99, write p99, guard p95 and accepted-age p95 where available. A transition
from busy to stale is NOT a rescue. No adoption gate is invented for this
diagnostic, and prior latency/staleness requirements remain unchanged.

First run small read-only/writer queued cells under race, then9 non-race arms.
The immediate arm is rerun on the SAME shared semaphore implementation; do not
compare only against old mutex numbers. Exclusive-create JSONL with sources,
dependency/runtime metadata and per-arm fsync. Preserve all failures. No
production configuration or service implementation changes.
