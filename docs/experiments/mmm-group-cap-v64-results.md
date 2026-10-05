# Guarded group scheduling v64 results

## Verdict

All three active caps FAIL the frozen joint screen in all three trials.
Caps1/2 shorten individual callbacks but increase transaction count, observation
backlog, drops and scheduled writer tails. Cap4 retains full completion and is
the best of these active arms, but still fails writer non-harm against native.
Do not adopt smaller caps, drop more observations deliberately, enlarge the
deadline or reinterpret fewer admitted records as equivalent work.

## Admission-boundary audit

SourceOwner.Admit holds its owner mutex while resolving service identities and
allocating monotonically ordered learner IDs. Durable.admitBatchResolved then
holds its durable mutex, validates order/capacity/retries, sets stopped=true,
calls the owned worker's actual Predict and Record, serializes those originals,
and atomically appends them. Only checked acknowledgments clear stopped. An
error/panic after staging requires close/replay, even for known rollback.

Consequently, moving this whole operation outside the service guard would not
merely move JSON preparation. It would change forecast timing, validity-through-
admission and recovery semantics. Current SourceOwner explicitly requires the
caller to hold the service guard. SQLite atomic batch and unique-source checks
do not independently validate a concurrently changing service snapshot.
Existing post-guard cleanup is different: it verifies already committed immutable
originals and appends an unlabeled terminal, not a new admission or feedback.

Pure immutable input copying/validation or source-resolution preflight may be
movable, but reservations must be rechecked under owner serialization, and lock
ordering must not invert service-guard -> source-owner -> durable-owner. A new
staging/commit API needs explicit cancellation, epoch, retry, uncertain-commit
and replay tests before it can replace the existing contract. No such API was
silently introduced by this experiment.

## Measured evidence

[Frozen protocol](mmm-group-cap-v64-protocol.md),
[raw JSONL](mmm-group-cap-v64.jsonl), SHA256:
`8757bd5eb90c8dabb29249863612fa5a5b2138f4164a2bfe868828b94de98583`.
All365 captured source hashes match embedded and local files. Go1.27.1,
Apple M4, darwin/arm64, GOMAXPROCS10. Twelve cells;2304 reads and1152 writes;
all72500 admitted originals and72500 terminals verify. No identity/phase errors
or guard-entry expiry. Queue-full drops are real and retained below.

Focused group-cap, old writer-arm and offered-arrival accounting race tests
PASS13.920s; service vet PASS. Experiment PASS19.739s means integrity/accounting,
not performance. Default fixtures retain the old cap; overrides are test-only.

Times in ms, nearest-rank percentiles. Scheduled writer tails include producer
lateness. Age is enqueue-to-observation completion for accepted observations;
it excludes dropped observations and therefore cannot excuse them.

| Trial | Arm | Accepted /192 | Dropped | Write p99 | Age p95 | Callback p95 | Groups |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 0 | Native four | n/a | n/a | 97.648 | n/a | n/a | n/a |
| 0 | Cap4 | 192 | 0 | 188.859 | 75.479 | 12.538 | 49 |
| 0 | Cap2 | 165 | 27 | 196.466 | 482.123 | 8.630 | 83 |
| 0 | Cap1 | 128 | 64 | 343.318 | 744.762 | 7.012 | 128 |
| 1 | Native four | n/a | n/a | 167.078 | n/a | n/a | n/a |
| 1 | Cap4 | 192 | 0 | 230.379 | 55.254 | 13.342 | 51 |
| 1 | Cap2 | 159 | 33 | 295.452 | 509.491 | 8.817 | 80 |
| 1 | Cap1 | 127 | 65 | 344.660 | 741.452 | 7.623 | 127 |
| 2 | Native four | n/a | n/a | 132.381 | n/a | n/a | n/a |
| 2 | Cap4 | 192 | 0 | 187.239 | 95.601 | 12.838 | 49 |
| 2 | Cap2 | 168 | 24 | 254.394 | 491.094 | 9.031 | 85 |
| 2 | Cap1 | 127 | 65 | 345.239 | 731.237 | 7.619 | 127 |

All active scheduled read p99 values remain24.250-48.257ms versus native
487.105-638.720ms. That read benefit does not cancel writer or observation harm.
The cap4 comparison fails the1.10 writer ratio in all three trials, even though
completion/age pass. Smaller caps also fail completion and250ms age criteria.

## Next lead

This rules out fixed smaller groups as a rescue for this workload, not every
scheduling policy. Retain batching and investigate pure preflight/copy work that
can be prepared without forecasting, allocating identities, holding an owner
lock across gate acquisition, or acknowledging before durability. First measure
that portion separately: if too small, prefer an explicit historical-admission
certificate design review over a cosmetic movement of the existing guard.
No runtime/default, paper claim, production or dependency change; all research
directions remain incomplete.
