# Guarded durable feedback feasibility v2

Frozen 2026-10-01 after v1's fixed-time contradiction and before changing the
test. Preserve v1's negative result. Use the same persistent service,
publication wrapper, bound SQLite learner, 32 explicit labels, replay checks,
unresolved-record deletion and cross-epoch negative control. The sole design
correction is chronological query time: query i occurs at `now+2i seconds`,
its feedback becomes available at `now+(2i+1) seconds`, and every new query
follows the previous acknowledged feedback. The pending query is at
`now+64s`. The future-only event is at `now+1h`, after every declared query.

Do not weaken model as-of checks or admission guards. Each admission must
validate the actual stored bound original while holding the service guard;
each feedback must freshly validate that original before a durable terminal
write. Future-only ingestion must preserve old-query compatibility. A visible
delete must reject the pending feedback before callback entry; discard it
without learning. Same-epoch replay must recover 32 labels and zero pending,
and new-epoch replay of the old log must fail. Measure replay time separately.

Race tests, a fresh independent source/identity audit, and package vet are
required. This finite sequence is not a full durable bridge, crash/power-loss
test, mixed-write load, automatic cross-epoch transfer or serving latency gate.
