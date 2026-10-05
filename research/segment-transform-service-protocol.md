# Actual shadow scheduler diagnostic

Three alternating off/on pairs. Each arm uses a new temporary persistent
LibraVDB service with50 fixture events, hash embedding32, recall50/pack10.
Four readers each issue16 sequential recalls; one writer inserts16 events at
2ms spacing. Requests share a frozen as-of; writes are later than it. Require
write/read overlap. Shadow policy capacity16, max age100ms, ordinary snapshot
compatibility (no temporal reuse). GOMAXPROCS4.

The processor calls the actual context-aware research segment fitter on its
fixed272-frame fixture, cap64. It deliberately does not infer labels from
retrieval scores. This tests processor/scheduler interference and accounting,
NOT real task learning or correctness/usefulness of learned answers. Completed
means diagnostic completion before snapshot/age invalidation, not verified
improvement. Immutable scalar output has no serving authority.

Freeze screen: every pair requires zero read/write errors, overlap, full terminal
accounting, on p99<=1.10*off p99, and at least52/64 completed shadow jobs (80%).
Drops/stale jobs count against completion. Nearest-rank p99 over64 requests is
the maximum; no population tail guarantee is claimed. Drain to terminal before
close, bounded5s wait. Record all request durations, status and source hashes
to exclusive-create JSON. No production, network or private data access.
