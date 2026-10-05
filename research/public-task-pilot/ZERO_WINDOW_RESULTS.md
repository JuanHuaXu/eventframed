# Zero coalescing delay: rescue rejected

## What changed

The isolated local dependency overlay changes exactly one initialization:
groupCommitWindow from1ms to0. The adaptive wait returns0 for this setting.
WAL flush, file.Sync, acknowledgement order, durability mode, index configuration,
admission and deadlines remain unchanged. Go's module-cache overlay restriction
required a copied module and research-zero-window.mod/sum. Normal go.mod/sum,
installed modules and production were not changed. go list confirms the build
resolves the copied module path.

Before workload execution, the selected race-enabled dependency tests passed:
synchronous durability defaults, group-sync acknowledgement, sync-failure
propagation and batch WAL recovery. Some tests set their own grouping windows;
they test storage invariants, not the candidate's performance advantage.

## Outcome

All32 arms completed, with1024 reads and512 writes. Every lower-rate arm passes;
every high-rate arm fails. The high-rate aggregates are:

| Admission | Original read errors /256 | Zero-window read errors /256 | Original write errors /128 | Zero-window write errors /128 |
| --- | --- | --- | --- | --- |
| off | 137 | 151 | 32 | 36 |
| on | 110 | 116 | 79 | 82 |

These are separate executions, so small count differences are not evidence of
a statistically established regression. They do clearly fail the predeclared
full admitted-grid rescue criterion. Eliminating this wait alone is insufficient.
There are385 total errors and93 operations canceled before callback entry.

All757 successfully returned journals and394 acknowledged writes survive orderly
close/reopen. Warmups pass, successful frontiers have no observed future evidence,
and the verifier checks source hashes and the exact one-line dependency change.
These successes do not override the latency failure, and they do not prove
arbitrary crash or failed-write rollback safety.

## Interpretation

The previous boundary measurements did not isolate pure coalescing time. This
ablation rejects the proposed sufficient rescue; it does not prove that the
delay costs nothing. Actual synchronous I/O, transaction/index work, request
serialization and the remaining read computation are still present. Removing
coalescing could also alter flush frequency; that was not measured, so no causal
claim about flush contention or power use follows.

Do not promote zero-window configuration or disable durability. Next profile
the uninstrumented read remainder and decompose the transaction path before
choosing batching, serialization reduction or another persistence experiment.
All seven full research goals remain open.

## Reproduction

ZERO_WINDOW_PROTOCOL.md preceded dispatch. Raw artifact: zero-window-results.json;
retained fresh databases: zero-window-results.json.stores.

```sh
node research/public-task-pilot/check-zero-window.mjs
go test -race -modfile=research-zero-window.mod -overlay research/public-task-pilot/zero-window-local-overlay-v1/overlay.json github.com/xDarkicex/libravdb/internal/storage/singlefile -run 'TestWALSyncDefaultsOnAndUnsafeModeIsExplicit|TestGroupCommitSyncsOnceAndAcknowledgesAllWriters|TestGroupCommitReturnsSyncFailureToWriter|TestBatchWALCommitRecoveryReplaysAllRecords' -count=1 -timeout=120s
```

The unused shared-cache overlay is retained as an unsuccessful setup artifact;
only zero-window-local-overlay-v1 was executable. Research dependency copies,
alternate module files and database files are not intended for source publication.
