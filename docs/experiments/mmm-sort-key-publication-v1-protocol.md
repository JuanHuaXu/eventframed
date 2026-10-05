# Sort-key publication v1: frozen private crash/reopen gate

The [migration probe](mmm-sort-key-migration-v1-results.md) showed that
ALTER/backfill can work, but record versions change without advancing
eventframed's runtime snapshot. This test-only prototype keeps serving
disabled until a durable SQLite sidecar publishes a READY marker bound to
the exact latest LibraVDB LSN, expected row count and a digest of the
durable collection. Production stores, sidecars and daemon are untouched.

Begin with a private legacy collection of three subsecond events and a
SQLite marker in PENDING. Simulate close/reopen after each boundary:
(1) pending marker, before ALTER; (2) ALTER plus one backfilled row;
(3) all rows backfilled and fully scanned, but before READY publication;
(4) READY publication. At the first three reopen points, the candidate
gate must deny serving. Publish READY only if a full scan verifies the
declared sort-key schema, exact key for every durable EventFrame, row
count, and identical LibraVDB latest LSN before/after the scan. Persist
that LSN, count and a stable row digest in SQLite. On reopen, verify the
digest once before accepting the marker. A hot read obtains LibraVDB LSN,
reads the SQLite marker, obtains LSN again, and admits only if both LSNs
equal the READY marker and the in-memory verified digest matches. The
accepted exact-LSN query at `.120Z` must return `.100Z` and `.120Z`
rows, not future `.125Z`.

After READY, perform an ordinary legacy write lacking the new key. The
new LSN must make the marker stale and serving must fail closed. A full
rescan must refuse to republish while that row remains unkeyed. An absent
sidecar must also deny rather than auto-authorize. Repeat the finite
state test under `-race`. Measure 1,000 quiet ready-gate checks (including
two LSN reads and one SQLite row read) and report p50/p99; frozen isolated
component ceiling p99 <=1 ms. No timing from race instrumentation is used.

A pass is only a private-store crash/reopen and gate-cost component. This
does not prove power-loss durability, concurrent multi-process writers,
an incremental marker update for every authorized write, loaded Recall
latency or Goal 6 completion. Any pending/ready exception or sidecar/LSN
alias is a failure, not an opportunity to loosen the gate on this cohort.
