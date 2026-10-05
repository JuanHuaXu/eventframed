# Verified source cleanup v54

PARTIAL RESCUE; the all-trial age criterion still FAILS. Combined cleanup reaches
570/576 completions but only one of three age p95 values is at most250ms. No
threshold was changed, unsuccessful trial dropped, or default path enabled.

## Contract and confirmed bug

`VerifyAndDiscardBatch` holds the private owner lock while reading canonical
stored originals by source, comparing complete expected records and committing
the existing atomic unlabeled terminal batch. Expected learner IDs are assertions;
terminal IDs come from validated storage. A late missing/mismatched member leaves
the whole batch undiscarded. Original timestamp instants must match; monotonic
clock data and location pointers are not persisted identity. Every other original
field is compared exactly. This grants no feedback or current service authority.

Audit found a separate confirmed defect in shared batch-discard retry handling.
Preflight accepted equivalent Available instants with Time.Equal, then marshaled
the caller's timezone representation into new bytes. The ledger correctly
rejected those bytes as a conflicting retry and the owner stopped. A regression
reproduced this in all four prepared/unprepared and point/snapshot-preflight
combinations; single-item discard was the passing control.

The fix retains the already-validated original terminal payload for retries.
Mixed new/retry batches and reopen now preserve those bytes. This is a lifecycle
correctness fix, not a performance intervention: the load fixture has only fresh
terminal writes. No existing stored data was rewritten or repaired.

## Verification

- Combined cleanup unit tests passed three race repetitions: source/field/time
  mismatch, duplicate/missing/nil-binding/canceled/oversized requests, no partial
  state or terminal, mixed retry/reopen, monotonic timestamp roundtrip, concurrent
  retries and error/panic recovery before/after actual commit.
- Existing useful feedback cannot be replaced by discard. The label in this
  test uses private test-only access, not a newly exposed evidence API.
- The timezone regression failed before the patch in all four batch variants,
  then passed with original-byte preservation. The combined API also passes an
  equivalent-instant retry with a different timezone representation.
- Targeted source/control/union/combined service race tests passed three repeats.
- Full ledger, learner and service race suites passed after the retry fix;
  their `go vet` and whitespace checks passed.

## Loaded result

Twelve rotated cells: off, raw prepared durable, batch-source/two-step cleanup,
batch-source/combined cleanup, over three trials. Each cell uses192 real Recall
requests and96 concurrent future-dated Observe writes on isolated LibraVDB/SQLite.
All guard/queue/frontier/durability settings remain unchanged. Cold public fixtures
only: no private data, labels, fitting or served forecast influence.

| Trial | Raw / two-step / combined completed | Combined drops | Two-step / combined age p95 ms | Off / combined read p99 ms | Off / combined write p99 ms |
|---|---|---|---|---|---|
| 0 | 192 / 178 / 187 | 5 | 311.372 / 265.888 | 30.999 / 31.400 | 18.762 / 31.627 |
| 1 | 192 / 177 / 191 | 1 | 329.522 / 259.033 | 33.015 / 28.984 | 18.528 / 28.971 |
| 2 | 192 / 178 / 192 | 0 | 328.032 / 240.460 | 36.956 / 26.424 | 17.874 / 34.100 |

Combined completes570/576 (98.96%) versus533/576 for two-step and576/576 raw.
Queue drops fall from43 to6; there is no entry expiry, unexpected error or
original mismatch. Ages cover accepted observations, not dropped work. All
28,500 combined,26,650 two-step and28,800 raw originals received full persisted
readback and durable unlabeled cleanup. Dropped observations are not labels.

Combined passes all three1.10x-off read p99 screens and one age screen. Raw age
p95 is155.945,201.037 and246.891ms. Combined write p99 remains above off in
every trial, although below raw write p99 (32.152,33.933,34.780ms) in this run.
This is not a population or zero-writer-overhead claim.

Mean combined post-guard time is7.001,7.120,6.697ms versus two-step9.105,
10.376,10.116ms. `VerifiedDiscardNS` includes actual readback, comparison and
terminal work in one span. Separate DurableVerifyNS/DiscardNS values are zero
because those pieces were not independently timed, not because work was omitted.
Containment and no-double-counting checks passed. The meaningful cross-arm
comparison is complete PostGuardNS, not the zero component fields.

Combined admission means remain6.009,6.032,6.107ms versus raw4.923,4.710,
5.043ms. Group sizes are recorded, so these are observed group means rather
than matched per-record causal effects. Source identity checks remain a necessary
contract absent from the faster raw control.

## Remaining lead

The source owner resolves source identity, after which durable admission preflight
performs a second learner-ID lookup. Investigate whether validated source results
can be reused under exclusive owner/ledger ownership without weakening actual
atomic append conflict checks, original forecasts or stop-on-uncertainty behavior.
This needs an explicit proof boundary and regression tests, not a blanket removal
of duplicate-looking reads or a public caller-supplied cache hint. No such rescue
is implemented or validated by this experiment.

Feedback/history authority, broader traffic and warm loaded learning remain
independently open. All seven research directions stay in progress.

Command: `EVENTFRAME_VERIFIED_CLEANUP_ARTIFACT=.../mmm-verified-cleanup-v54.jsonl go test ./internal/service -run '^TestResearchVerifiedCleanupExperiment$' -count=1 -v`.
Completed in20.96s on Go1.27.1 darwin/arm64, Apple M4, GOMAXPROCS10. All12 cells
passed accounting/integrity; source hashes, counters and combined-phase
containment were independently checked from JSONL.

SHA-256:
`9155b251d2c1c6deab5f09524b72feeb0622109b070534a0b1e5be41219381e9`.
