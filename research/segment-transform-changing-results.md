# Changed-work controls: conditional benefit, broad failure

**FAIL all six frozen completion cells.** Exact reuse helps repeated groups,
but never reaches52/64 under this changing-work schedule. It cannot accelerate
all-distinct inputs. Earlier identical-work success remains scoped accordingly.

| Work | Trial | Direct completed | Memo completed | Memo hits | Successful fits |
|---|---:|---:|---:|---:|---:|
|Four groups|0|1|25|23|2|
|Four groups|1|1|25|23|2|
|Four groups|2|0|47|43|4|
|All distinct|0|1|1|0|1|
|All distinct|1|1|0|0|1|
|All distinct|2|1|1|0|1|

All completion denominators64, including drops and stale outcomes. One distinct
fit returned successfully but failed the scheduler's subsequent freshness check;
successful fits must not be equated with completed jobs. No wrong-history
forecast, read/write error or accounting failure occurred. Every returned value
was compared against its own independently precomputed reference before return.
The precomputed references never populate the cache or supply returned values.
All-distinct hit counts are exactly zero, so no false cross-work acceleration.

Memo p99 passed the1.10x shadow-off threshold in all six cells. Against direct
fitting, five passed; distinct trial0 failed (32.546ms versus27.621ms,1.178x).
All16 writes overlapped readers in every arm. No general foreground non-harm
guarantee follows from these18 small arm measurements.

Artifact: `segment-transform-changing-results.json` includes1152 recall and288
write timings, terminal counts, fits/hits/verified returns and source hashes.
`node research/verify-segment-transform-changing.mjs` verifies all18 arms and
reports `passed:false`. Actual service/storage/scheduler are used, but labelled
workloads remain controlled fixtures; this is not real-task learning quality.

## Research implication

Exact caching is a supported optimization for repeated snapshots, not a rescue
for fresh distinct evidence. Enlarging the cache cannot help the zero-hit arm.
Deadline-aware admission could avoid wasted compute but cannot honestly count
rejected work as completed. The next useful compute lead is an incremental
update or cheaper declared challenger for changed evidence, evaluated against
the exact forecast and original quality criteria. Prior MAP/VB/other challenger
failures remain relevant; do not claim speed alone rescues their accuracy.

For this full segment model the quadratic interval recurrence still dominates.
A research incremental method must preserve delayed-label chronology, eviction
and hazard semantics, or explicitly declare approximation error. This is a
different problem from sharing unrelated histories, which remains forbidden.
Direction6 and the other six directions remain open. Production unchanged.
