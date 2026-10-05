# Bound worker v8: origin-bound as-of rebinding

Status: frozen before implementation; local research screen run 2026-10-01.
Production unchanged.
Motivation: [v7 diagnostic](mmm-bound-worker-motion-v7-results.md).

An opt-in motion-mode epoch has two distinct identities:

- **Origin snapshot:** the exact event-store snapshot where the epoch was first
  bootstrapped. Store its canonical representation atomically with the seal.
- **Current target:** the snapshot against which source histories are freshly
  validated on each open. It is not included in the motion-mode HMAC.

The motion-mode HMAC has a new domain separator and commits to origin,
tenant, stream, epoch, seed, cutoff, key ID, every cutoff-eligible verified
label pair, and every source-validator retain decision. Discards and
unresolved predictions are parsed for source-log validity but do not train the
rebuilt model or enter this seal; the unchanged-sequence guard excludes a
concurrent source-log append during final handoff.
The legacy exact-target HMAC and ledger open path stay unchanged. A log with a
legacy seal but no origin metadata cannot be silently upgraded to motion mode.

On every open, before the worker is exposed:

1. Read and validate the stored origin, if present; on first open, use the
   current target as origin. Rebuild components only from the complete source
   log, validating sources against the current target.
2. Require the publisher's bounded as-of compatibility proof from origin to
   current target at the bootstrap cutoff. Missing history, backfill or a
   general mutation rejects, even if every source event is byte-identical.
3. Recompute the origin-bound HMAC. The ledger atomically accepts only the
   same seal and origin. The first bind requires an empty log.
4. Replay this epoch's canonical original forecasts, requiring a binding,
   witness, source continuity, and as-of compatibility from each original's
   own admission snapshot at its prediction time. Never recompute forecasts
   after outcomes. A wrong key or changed retained decision rejects.
5. Privately start the worker, then hold the exact current store guard and
   unchanged source-log sequence guard for final certification. Failure closes
   the unexposed worker; a sealed unused research log may remain.

Tests: memory and persistent stores; future-only ingestion then orderly
restart with exact source retry and next-ID continuation; backfill after that
restart; legacy log rejection; wrong origin, seal, key, changed retained label
and missing motion proof. Benchmark rebuild and reopen separately. This is not
a power-loss, loaded p99, or real-agent answer-quality claim.
