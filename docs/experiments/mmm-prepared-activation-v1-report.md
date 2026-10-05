# Prepared acceptance feasibility, exploratory v1

## Decision

The two-phase design is **not promoted**. A bounded interleaving model found a
safe ordering *under its atomic-step assumptions*, but the test-only SQLite
implementation did not deliver a convincing performance rescue. At 200 records,
mean guarded work fell only from 2.671 to 2.449 ms while mean total write work
rose from 2.671 to 3.845 ms. No service-load or agent outcome was measured.
Goal 6 remains open, and production is unchanged.

This protocol and report were written after the exploratory run. The timing
numbers are therefore diagnostic, **not** untouched confirmation against a
predeclared gate. The first timing artifact, `mmm-prepared-activation-v1.jsonl`,
preceded a test-control correction and is retained but superseded by
[the final raw run](mmm-prepared-activation-v1-final.jsonl), SHA-256
`20fb3cebcf6479edce4c54030473b6fda9ca4a5542ac3b3a2ec696a318e45753`.
The final test source SHA-256 is
`f40d99745ddc9776550eb0ca4fd927668e3641cbe80002649dea1392f71876b9`;
the model source SHA-256 is
`beb580cb70d56870ee0f0b2e51101de4a71645c496e7b75661cbe626b593f6e0`.

## Mechanism and safety

[The model](../../research/prepared-activation-interleavings.mjs) enumerates
one prepared admission, one incompatible publication, one competing admission
for either the same or a different source, and an optional crash/reopen.
Prepared rows are invisible; activation atomically creates the accepted marker
and unique source/ID references. A source/version/ID check stays under the
publication guard until that activation commits. The model assumes the
accepted-marker transaction is atomic and that publication obeys the guard;
it does not model SQL, mutable worker fits, process locks, power loss or time.

| Model arm | Different-source schedules / invalid | Same-source schedules / invalid |
| --- | ---: | ---: |
| Guard through activation and check provisional ID | 72 / 0 | 64 / 0 |
| Release guard before activation | 94 / 16 | 83 / 9 |
| Omit provisional-ID check | 78 / 8 | 64 / 0 |

The unsafe controls expose stale-version commits, duplicate IDs and false
acknowledgments. The guarded model also exercises accepted and rejected paths,
orphaned preparations, and exact-byte replay after an uncertain acknowledgment.
These finite counts are exhaustive only for the stated small model, not a
concurrency theorem for the daemon.

[The isolated SQLite test](../../internal/researchledger/prepared_activation_experiment_test.go)
stores each prepared payload under WAL/FULL outside the modeled guard. Under the
guard it checks the captured version and sequential ID, then inserts per-source
unique references plus acceptance in one transaction. Accepted lookup joins
the reference to its prepared original. The unchanged control is the existing
`AppendBatchPrepared` with its service identity index. The candidate's prep
identity parser is test-only; neither actual service-source validation nor
worker forecasting is executed.

The semantic test passes with `-race`: preparation is invisible across reopen,
stale activation rejects, committed originals replay exact bytes after reopen,
exact retries resolve the old original, changed retries conflict, and a late
duplicate-source conflict rolls back the whole batch. The complete ledger race
suite and `go vet ./internal/researchledger` pass. These are orderly reopen
checks, not process-exit or hardware power-loss tests.

## Timing and scope

Three rotated trials per size/arm, 32 fresh batches per cell, 1,024-byte payload
padding, Apple M4, Go 1.27.1, local SQLite WAL/FULL. Preparation and activation
are measured separately. The control's entire append is guarded. All 48,000
originals across both arms were checked by source and exact bytes after reopen.
Each summary below pools 96 batch samples per size/arm; milliseconds.

| Records | Arm | Guard mean | Guard p95 | Total mean | Total p95 |
| ---: | --- | ---: | ---: | ---: | ---: |
| 50 | Existing indexed append | 0.855 | 1.235 | 0.855 | 1.235 |
| 50 | Prepare then activate | 0.678 | 0.815 | 1.098 | 1.215 |
| 200 | Existing indexed append | 2.671 | 6.018 | 2.671 | 6.018 |
| 200 | Prepare then activate | 2.449 | 2.680 | 3.845 | 7.057 |

At 50/200 records, guarded means improve about 20.7%/8.3%, while total means
worsen about 28.4%/44.0%. The 200-record guarded median is slightly worse
(2.415 versus 2.446 ms); the p95 decrease is driven by a smaller tail in this
short, unloaded run and is not a loaded p99 result. The control and candidate
have different schemas, so this does not isolate a single SQL operation.
Neither arm includes the real guard acquisition, source reads, worker staging,
feedback, concurrent writes, crash recovery, or full-request latency.

Reproduce the model and checks:

```sh
node research/prepared-activation-interleavings.mjs
go test -race ./internal/researchledger -run '^TestPreparedActivationSemantics$' -count=1
go test -race ./internal/researchledger -count=1
go vet ./internal/researchledger
```

The opt-in timing test requires a new exclusive artifact path:

```sh
EVENTFRAME_PREPARED_ACTIVATION_ARTIFACT=/absolute/new/path.jsonl \
  go test ./internal/researchledger -run '^TestPreparedActivationExperiment$' -count=1 -v
```

## Next boundary

Do not install this two-phase schema or infer that the service's writer tails
would improve. To revisit staged acceptance, first identify a materially
cheaper *identity-restored* activation transaction or reduce per-source work
without sacrificing uniqueness, exact retry, order, or durable acknowledgment.
Then test the real store-before-owner lock order, stale model snapshots,
cross-process ownership, orphan cleanup, crash points, and loaded service
latency with the original non-harm gate. The prior envelope and exclusive-lock
failures remain negative; none is erased by this feasibility model.
