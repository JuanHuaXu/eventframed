# Opt-in source batch reads v53

PARTIAL RESCUE, but FAIL under the unchanged loaded completion-age criterion.
Batch source reads improve capacity substantially; all three age p95s still
exceed250ms. No whole research direction is completed and no default changed.

## Owner integration and checks

`OpenSourceOwnerBatchReads` is an explicit opt-in variant. It batches logical
source resolution for admission and discard; `LookupBatch` reads full originals
in caller order from one snapshot. Default constructor/single reads remain the
control. Actual forecasts and durable writes are unchanged. No caller learner ID
enters source admission, and no feedback or serving authority is exposed.

The owner checks response count/source binding/missing shape/payload budgets,
then canonical forecast encoding, model contract, seed, epoch, ledger identity
and complete service binding. Missing and malformed records remain different;
readback requires every requested original. Source retry comparisons and atomic
terminal handling remain in place. Cancellation after decode is checked before
return; uncertain writes still stop the owner until replay.

Tests passed:

- Three race repetitions of all source-owner tests, including the batch variant.
- Identical warm training histories with per-label publication barriers preserve
  pre-outcome originals; warm batch readback and constructor swaps on reopen
  preserve complete original records, not freshly recomputed forecasts.
- Duplicate/conflicting admission, missing readback/cleanup, cancellation, output
  aliasing and count bounds fail before partial state changes.
- Malformed count/source/miss/sequence/kind/stream/canonical/seed/epoch/binding/ID
  and oversized response cases reject instead of silently authorizing a record.
- Before/after actual commit acknowledgment failures for admission and discard
  stop the owner; reopen resolves new versus exact-retry outcomes correctly.
- Source/control/union/guard load accounting passed three race repetitions.
  Full ledger, learner and service race suites and their `go vet` passed.

Warm labels in the unit test use private test-only access and a fixed matched
schedule. They do not provide a public trusted-feedback mechanism or validate
general schedule-invariant learning.

## Loaded result

Twelve rotated cells: off, prepared raw durable, point-source owner, batch-source
owner, over three trials. Each has192 Recall requests and96 future-dated Observe
writes in isolated LibraVDB/SQLite. Queue64, four readers, K50/pack10, groups<=4,
20ms guard-entry budget, FULL writes and the250ms age limit remain unchanged.
All data are synthetic public fixtures. No fitting/feedback/production access.

| Trial | Raw / point / batch completed | Batch drops | Point / batch age p95 ms | Off / batch read p99 ms | Off / batch write p99 ms |
|---|---|---|---|---|---|
| 0 | 192 / 133 / 182 | 10 | 578.970 / 303.270 | 31.463 / 26.786 | 19.009 / 36.329 |
| 1 | 192 / 134 / 177 | 15 | 569.746 / 320.670 | 33.944 / 27.803 | 17.748 / 35.000 |
| 2 | 192 / 133 / 180 | 12 | 587.489 / 308.006 | 31.942 / 23.526 | 18.833 / 26.032 |

Batch source completes539/576 observations versus400/576 for point source and
576/576 raw. Drops fall from176 to37, with no entry expiry or unexpected error.
The age distributions cover accepted observations only, not dropped work. Every
accepted original was verified and durably discarded:26,950 batch-source,
20,000 point-source and28,800 raw records. Serving-read p99 passes the1.10x-off
screen in all batch trials; age fails in all three. Raw age p95 is221.836,
192.032 and195.764ms, passing the finite age screen throughout.

Batch write p99 remains above off in every trial. Compared with raw write p99
(34.067,30.879,31.961ms), batch worsens trials0/1 and improves trial2. A passing
read-tail screen is not a broad zero-overhead or writer non-harm result.

Accepted-group means: batch admission5.89-6.35ms versus point9.34-9.92ms;
batch readback2.87-3.04ms versus point6.95-7.81ms; batch discard6.82-6.93ms
versus point10.31-11.51ms. Batch post-guard work is9.70-9.97ms versus raw
6.02-6.74ms. Group sizes are recorded; these observed group means are not a
matched per-record causal effect estimate. Source and raw paths retain different
identity guarantees, so raw speed is a control rather than an acceptable bypass.

## Next lead

`researchFinishSourceGroup` reads and verifies originals, then calls source
`DiscardBatch`, which resolves and canonically decodes the same originals again.
This is useful evidence for a combined verified-discard operation under the
private owner lock: resolve once, compare complete expected originals, then
perform the existing atomic terminal batch. Any missing/mismatched/late-invalid
member must reject before a terminal write; uncertain commit must still stop.

This proposal must NOT skip actual persisted readback, accept caller learner IDs
as authority, mutate original forecasts, weaken durability or certify feedback.
Test its owner/lifecycle contract and compare with the retained two-step source
path under unchanged loaded screens. It is a live lead, not an already validated
rescue. Broader traffic and feedback/history authority remain independently open.

## Artifact

Command: `EVENTFRAME_SOURCE_BATCH_LOAD_ARTIFACT=.../mmm-source-batch-load-v53.jsonl go test ./internal/service -run '^TestResearchSourceBatchLoadExperiment$' -count=1 -v`.
The test completed in20.84s on Go1.27.1 darwin/arm64, Apple M4, GOMAXPROCS10.
All12 cells passed accounting/integrity; source hashes and counters were
independently recomputed. No sample/trial was excluded or threshold retuned.

SHA-256:
`d209db33b6608e33061d3f5cf9d7a345773b234d05c8f20b310789a23a4efef7`.
