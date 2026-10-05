# Guarded durable feedback feasibility v1

Frozen 2026-10-01 before the cross-layer run. This is an isolated feasibility
test, not a new service bridge or an adopted persistence path.

Use one temporary persistent LibraVDB service, the research publication
wrapper, one fixed as-of query time, and one exclusively owned SQLite learner
log. Issue 32 separate journaled recalls against the same visible public
fixture event. Every ledger admission must occur inside the service's
as-of admission guard with the actual journal/event/feature/baseline binding.
Every externally supplied feedback label must separately pass a fresh service
guard before `Durable.Feedback` writes its terminal record. After the first
16 labels, add a future-only event; subsequent guarded feedback at the old
query time must remain valid. All labels are explicit fixture values, never
inferred from rankings or absent replies.

Close and reopen the durable learner. A same-epoch replay must recover all
32 labels, zero failed/pending updates, original bound records, and finite
scores. Then admit one more bound record but do not label it. Delete the
visible source event through the service. The fresh guard must reject a
feedback attempt before its callback runs; discard the unresolved durable
record without turning it into a negative label. Reopen again: 32 labels,
zero pending and a recorded discard must remain. The old bound record must
not acquire service authority from successful learner replay.

Run under the race detector. Record replay cost separately from service
latency. This test uses one query horizon and no store/process crash. It
cannot authorize cross-epoch model transfer, arbitrary mixed-mutation
learning, automatic durable bridge startup, or loaded p95/p99 claims.
