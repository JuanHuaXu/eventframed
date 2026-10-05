# Bound worker v6: guarded continuing admissions

Date: 2026-10-01. Protocol: [v6 contract](mmm-bound-worker-v6-contract.md).
Status: partial component validation; no production or served-law adoption.

## What passed

The service now returns an owned research worker, not the raw durable journal.
Before an admission exists, a separate candidate validator checks the actual
committed query journal, source event, baseline and extracted 5W1H features.
Source continuity to the current snapshot is checked before taking the as-of
guard, then the exact snapshot is rechecked under the guard. The callback
constructs a keyed source witness and journals the **actual** forecast.

The bound SQLite ledger enables a unique `(tenant, stream, journal, event)`
source index. A same-source retry returns the original forecast and learner
ID after restart; changed features reject. Replay requires bound, witnessed
originals and validates each against the current source before the worker is
exposed. Verified feedback requires a fresh source-continuity proof, request
and journal check, and guard; a missing outcome stays pending.

The memory and persistent-store fixtures each exercised guarded open, a new
journaled admission and label, exact retry, restart, the next learner ID,
restart with an unlabeled admission, wrong-query rejection, and rejection
after deleting the source event. Stale source-log or event-store preparation
also rejected. The direct bound-ledger fixture additionally rejected a wrong
witness key, changed transferred outcome, cold replay and a copied bootstrap
handle attempting to start a second worker.

## Failed first design

The first feedback design nested `ValidateResearchTransferSource` inside the
store's as-of guard. Both backends returned `research event continuity guard
busy`: the continuity check acquires that same guard. The corrected design
proves continuity outside, then requires the exact target snapshot inside the
guard before writing. Restart replay likewise runs privately before a final
store-before-source certification. A failed final certification closes the
unexposed worker; it may leave a sealed, unused research log on disk.

## Cost and checks

Three isolated 50-iteration repeats on this Apple M4 measured guarded
admission at **0.141-0.183 ms/op** and guarded feedback at **0.130-0.135
ms/op**. The benchmark uses an in-memory event store, a cold bound epoch that
warms during the run, and one public fixture event. Journal creation, waiting
for background fitting, and the opposite operation are outside each timed
section. These numbers are not loaded p95/p99 or full-request latency.

```text
go test ./internal/researchledger ./internal/researchmemory ./internal/service -count=1
go test -race ./internal/researchledger ./internal/researchmemory ./internal/service -run 'TestBoundDurable|TestBoundTransfer|TestResearchSourceWitnessSurvivesUnrelatedDeletion|TestReplayBoundary|TestResearchDurableValidation' -count=1
go vet ./internal/researchledger ./internal/researchmemory ./internal/service
go test ./internal/service -run '^$' -bench '^BenchmarkResearchBoundWorkerGuardedOperations$' -benchtime=50x -count=3
```

All final checks passed.

## Still open

The low-level `OpenBoundDurable` remains an internal research primitive;
service-originating work uses the guarded handle. The epoch seal binds an
exact target snapshot, so future-only store motion across a later restart
needs a separately validated rebinding protocol. No loaded persistent-write
p99, crash/power-loss recovery, original-source pending transfer, scored-law
publication, or untouched agent-task improvement is established. Externally
supplied usefulness is still an assumption, not a model-generated label.
Goal 6 remains open.
