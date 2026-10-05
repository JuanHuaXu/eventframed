# Bound worker v10: fitted prediction after abrupt exit

Date: 2026-10-01. Frozen [v10 protocol](mmm-bound-worker-fit-exit-v10-protocol.md).
Research-only; production unchanged.

The finite fresh-process screen **passes**. One witnessed positive source
label and 31 verified negative worker labels reached the 32-sample fit
boundary. Before abrupt child exit, an unlabeled probe had baseline
`0.79911539` and fitted probability `0.62124220`, exceeding the frozen 0.01
effect threshold. After reopening persistent LibraVDB, durable publication
lineage, source log, and motion worker in a new process, the worker had 32
completed labels and the pending probe. A new committed query at the same
as-of time had identical features and baseline and produced the same
`0.62124220` probability by exact floating-point comparison. The next ID and
source index also survived. Deleting the source then blocked a further
motion handoff. The separate [v9 screen](mmm-bound-worker-exit-v9-results.md)
keeps the pre-cutoff backfill rejection control.

Reproduce with:

```sh
go test -race ./internal/service -run '^TestResearchMotionBoundWorker(Acknowledged|Fitted)AbruptExit$' -count=1 -v
go test ./internal/researchledger ./internal/researchmemory ./internal/researchpublicationstore ./internal/service -count=1
go vet ./internal/researchledger ./internal/researchmemory ./internal/researchpublicationstore ./internal/service
```

This is acknowledged-write **process-exit** recovery of a fitted research
model, not hardware power-loss, uncertain-commit recovery, cross-database
atomicity, population loaded p99, or an improvement in agent answers. The
worker remains opt-in and does not change served forecasts. Goal 6 OPEN.
