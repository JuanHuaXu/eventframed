# Bound worker v5: sealed continuing research epoch

Date: 2026-10-01. Scope: opt-in research components only; production unchanged.
Protocol: [v5 contract](mmm-bound-worker-v5-contract.md).

## Result

The local component screen passed. A keyed HMAC commits to the target snapshot,
tenant, destination stream, epoch, seed, cutoff, key ID, every eligible original
prediction and terminal record, and each source-validator retain decision. The
sealed model and commitment travel as one opaque value. An empty SQLite epoch
ledger binds the commitment before its first admission; a restart with the
same validated evidence replays this epoch's **original forecasts** on top of
the rebuilt component model. It does not invent negative labels for absent
feedback. The bootstrap is single-use, including copies of its public handle,
so one adapter cannot back two workers.

In the 64-label fixture, 32 surviving labels initialize epoch 2. One new
admission and verified feedback were learned, then the worker closed and
reopened from the same source evidence. Its score and completed count matched
the pre-close state, and the next prediction continued at ID 2. A changed
retained outcome, ordinary cold replay, and binding to a preexisting unsealed
log were rejected. Service-level memory and persistent-store fixtures also
passed guarded open/reopen; source-log and event-store mutations between
preparation and open rejected before handoff.

Checks run:

```text
go test ./internal/researchledger ./internal/researchmemory ./internal/service -count=1
go test -race ./internal/researchledger ./internal/researchmemory ./internal/service -run 'TestBoundDurable|TestBoundTransfer|TestResearchSourceWitnessSurvivesUnrelatedDeletion|TestReplayBoundary' -count=1
go vet ./internal/researchledger ./internal/researchmemory ./internal/service
```

All passed. On this Apple M4, three 100-iteration isolated benchmark repeats
measured 64-label sealed rebuild at 0.793-0.852 ms/op, and empty bound-ledger
reopen **including rebuild** at 1.172-1.195 ms/op. These are means under an
isolated benchmark, not p95/p99, queue age, power-loss recovery, or full
service-request latency.

## Limits

The service guard covers preparation and epoch opening, not subsequent use of
the returned research worker. Future admissions still need per-record service
source authorization, and a new epoch does not inherit unresolved predictions
or late feedback from the source epoch. A later source-log change may force a
new seal on restart. The continuing worker is not wired into served forecasts,
and no loaded write/concurrency, crash, or untouched agent-task result was
established. Goal 6 remains open.
