# Bound worker v8: origin-bound as-of rebinding

Date: 2026-10-01. Scope: opt-in research APIs in `eventframed`; production
serving and the legacy exact-target bound-worker path were not changed.
Frozen design: [v8 contract](mmm-bound-worker-motion-v8-contract.md).

## Result

The finite local screen **passes its stated orderly-restart controls** on
memory and persistent LibraVDB backends. A worker rebuilt from one witnessed,
verified source label admitted and learned one new guarded label. After an
event available only after the bootstrap cutoff was ingested, the same sealed
worker reopened with two completed labels, preserved exact source retries,
and continued allocating the next prediction ID. A backfilled event before
the original observation time then rejected preparation. The [v7 diagnostic](mmm-bound-worker-motion-v7-results.md)
shows why unchanged source identity alone would have been insufficient.

Negative controls also pass: a store without an as-of publication proof cannot
prepare the motion worker; a legacy sealed log with no origin cannot be
upgraded; changed origin or seal is rejected by the ledger; and changed key,
retained outcome, or retain decision changes the model seal. The new mode uses
an origin-bound HMAC domain distinct from the legacy exact-target domain.
Focused race checks, `go vet`, and full `researchledger`, `researchmemory`,
and `service` package suites passed.

## Component cost

Apple M4, Darwin arm64, 100 iterations per run, three runs:

| Isolated operation | Observed mean range |
| --- | ---: |
| Rebuild a 64-label motion seal, retaining 32 labels | 0.801-0.920 ms |
| Open/reopen an empty motion journal with a prebuilt seal | 0.370-0.383 ms |

The open benchmark excludes rebuild, live source validation, the service's
as-of proof, and Close. It is **not** a full handoff or loaded request latency
measurement. Rebuild is an offline operation, not a per-recall cost.

## Reproduction

```sh
go test -race ./internal/researchledger ./internal/researchmemory ./internal/service -run 'TestMotionBootstrap|TestMotionSeal|TestResearchMotionBoundWorker|TestResearchBoundWorkerMotionProofDiagnostic' -count=1
go test ./internal/researchledger ./internal/researchmemory ./internal/service -count=1
go vet ./internal/researchledger ./internal/researchmemory ./internal/service
go test ./internal/researchmemory -run '^$' -bench '^BenchmarkMotionBound' -benchtime=100x -count=3
```

Goal 6 remains **OPEN**. The controls prove an orderly, finite restart with
future-only motion, not power-loss recovery, population p99 under concurrent
writes, state recovery after a visible mutation, or better untouched agent
outcomes. No promotion or push follows from this component result.
