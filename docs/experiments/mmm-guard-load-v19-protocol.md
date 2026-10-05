# Exact-guard persistent admission load v19

Freeze before execution. Diagnostic only; no production-readiness pass gate.
Three rotated trials of off, validation-only guard, and durable guard. Each arm
has a fresh persistent LibraVDB store,50 visible fixture events,192 recalls from
four readers, and96 future-only writes spaced2ms apart. Recall50/pack10; enabled
arms use a64-observation queue. No private data or production service is used.

Every accepted observation validates ALL50 candidates under one exact-snapshot
mutation guard. The durable arm admits each original cold prediction with its
actual service binding, then durably discards it without a label. The validation
arm uses the same cold record construction/validation but no ledger. Guard busy
and stale rejections are separately counted, never silently retried. Queue drops
are separate. This is admission/abandonment infrastructure, NOT feedback fitting
or a full durable learning bridge. No labels or accuracy claim are invented.

Report raw request/write/guard durations, accepted-observation age, queue drops,
attempts, busy/stale rejections, full observations accepted, candidate validations,
durable admissions/discards and errors. Conservation: attempts+drops=192 for
enabled error-free arms; attempts=accepted+busy+stale;50 validations per accepted
observation. Durable admissions=discards=validations. Empty acceptance is a
measured availability failure, not a fast success.

Compare p99 Recall and write durations against same-trial off; no retrospective
threshold to turn low acceptance into a pass. Full worker learning, feedback
authority, restart history, unique service-event mapping and realistic task
outcomes remain separate work. Discard is not a label and needs no freshness
permission to abandon a record. Current fixture serializes it inside the guard,
so timings include that conservative critical-section placement.

Run a small8-recall accounting fixture under race first, then9 non-race load
arms. Exclusive-create JSONL with embedded source/protocol hashes and per-arm
fsync. Retain failures and reject incomplete artifacts. No source changes to
the guard or service implementation are part of this diagnostic.
