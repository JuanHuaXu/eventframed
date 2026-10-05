# Bound worker v6: guarded continuing admissions

Status: pre-implementation research contract. Production serving is unchanged.

Question: after a sealed v5 epoch opens, can its new observations be admitted
and labeled only through the service's source guards, with retry identity and
original forecasts preserved across restart?

Invariants:

1. The service exposes an owned research handle, not a raw Durable. Each
   candidate admission binds one committed journal, query, event, baseline,
   extracted feature vector, as-of time, tenant, epoch, and keyed source
   witness under the store's as-of guard. The durable source identity index
   reserves `(tenant, stream, journal, event)` across restarts.
2. An exact source retry returns the same original forecast and learner ID;
   changed features, baseline, time, binding or witness reject without a new
   label. Duplicate-source conflicts cannot be converted into new evidence.
3. A verified outcome is admitted only after fresh source/request validation
   under the store guard. Absent feedback remains absent; changed/early or
   out-of-order feedback rejects. The method does not authenticate external
   truth and cannot infer usefulness from its own forecast.
4. The bound replay accepts only witnessed, bound originals from this epoch.
   Original expert probabilities are replayed, never recomputed after labels.
   Existing v5 unbound research journals fail closed on v6 open.
5. Following an orderly restart from the same validated source history, an
   exact retry and new observation preserve counts and next-ID order. A source
   deletion or backfill invalidates guarded operations; future-only motion is
   allowed only when the existing as-of proof supports it.

Controls: memory and persistent stores, same-source retry, changed-input
retry, stale event-store snapshot, stale source-log preparation, restart,
missing feedback, and source mutation. Benchmark service admission and
feedback separately; report loaded p99 only after a loaded experiment.

This is a research admission contract, not a claim of agent improvement or
full Goal 6 completion. Power-loss, mixed writes, and scored-law publication
remain separate required tests.
