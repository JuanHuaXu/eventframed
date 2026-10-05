# Snapshot-bound research admission v17

## Question and invariant

The v16 final snapshot check rejects a mutation during validation, but does not
exclude a mutation between the final check and a later ledger commit. This is a
confirmed timing gap for consumers that treat validation as a reservation.

The research adapter now provides `WithResearchSnapshot`: acquire its mutation
mutex without waiting, require an exact current non-quarantined snapshot, and
hold the mutex through a bounded callback. The service's
`WithValidatedResearchAdmission` performs existing journal/query/event validation
inside that guard before entering the caller's ledger callback. Unsupported
stores, busy guards and stale snapshots fail closed.

This is an unconfigured research API, not a production serving change. All backend
mutation must pass through the exclusively owned adapter. A callback must not
mutate or close the adapter; doing so would deadlock. It must not retain access
as though the guard remained valid after return.

## Checks run

- Snapshot mismatch, nested/busy guard and cancelled context reject callbacks.
- Callback errors and panics release the mutation mutex.
- A concurrent policy writer does not finish during the guarded callback's
  20 ms observation window; it completes after release and changes the snapshot.
  The separate mutex-ownership assertion directly checks exclusion rather than
  relying only on scheduling timing.
- Actual Recall journal validation enters a durable SQLite admission callback;
  the committed record retains its service binding. A subsequent policy mutation
  prevents reuse of the old guarded admission. Native unsupported stores reject.
- The original two targeted tests passed three race repetitions.
- After adding the concurrent writer test, the complete publication-store,
  research-memory and service packages passed one race run. Vet passed for all
  three packages. Environment-gated load experiments were not rerun.

Commands:

```sh
go test -race ./internal/researchpublicationstore ./internal/service -run 'TestSnapshotGuardExcludesMutations|TestGuardedServiceAdmissionCommitsBoundLedger' -count=3
go test -race ./internal/researchpublicationstore ./internal/researchmemory ./internal/service -count=1
go vet ./internal/researchpublicationstore ./internal/researchmemory ./internal/service
```

## Result and remaining work

PASS for the tested snapshot exclusion and bound-admission primitive. This is
not a cross-database atomic transaction, full durable service bridge, feedback
authentication, or restored-history authority. Mutation after guard release
still requires fresh validation before later learning. Callback blocking extends
writer exclusion; loaded latency and contention must be measured before adoption.
The tests use memory-store guarded service admission with a real SQLite research
ledger; persistent-backend guarded service load remains untested.
