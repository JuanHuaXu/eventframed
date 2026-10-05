# Temporal reuse: bounded persistent screen PASSED

Same future-arrival workload as v3, now with opt-in temporal compatibility:

| Trial | Off p99 | On p99 | Completed / accepted | Stale |
| --- | --- | --- | --- | --- |
| 0 | 31.243 ms | 29.287 ms | 255/256 | 0 |
| 1 | 31.086 ms | 27.239 ms | 254/256 | 1 |
| 2 | 27.362 ms | 29.236 ms | 255/256 | 0 |

All1536 reads and384 writes succeeded. All64 writes per arm overlapped active
readers. Each enabled p99 satisfies the10% paired allowance, and every enabled
arm exceeds90% completed/accepted. Cancellation accounting balances after Close.
One stale job can include deadline/shutdown invalidation; it is not evidence of
an unaccounted semantic change. Per-request durations/statuses are retained in
mmm-shadow-temporal-v4.json. No retries to replace unfavorable trials.

Source-level gate: job carries AsOf; snapshot and full retained ingest history
are read under the store lock. Only fully accounted future ingestion can reuse
results, with runtime/evidence-epoch increments matching and all semantic
versions unchanged. Missing history, cutoff, optional interface, semantic motion
and unsafe version boundaries fail closed. No journal acceptance rule changed.

Unit/race checks cover future-only reuse, backfill rejection, cancellation,
semantic-version changes, missing history, zero cutoff, unsupported-store fallback
and shutdown conservation. Store/service vet passed. The snapshot result remains
a historical diagnostic, not a lease or authorization to affect served forecasts.

Compared with v3's39-95 completions per enabled trial, the recorded pattern is
consistent with avoiding irrelevant invalidations. Sequential cross-run timings
do not prove a causal speedup or a population p99 guarantee. This is one small
embedded-database workload; arbitrary backfills and real learned workers remain
outside the validated scope. TemporalReuse defaults false and production remains
unchanged. Roadmap6 is advanced, not complete.
