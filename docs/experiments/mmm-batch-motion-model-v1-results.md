# Batch motion authority v1: finite model passes

The frozen [protocol](mmm-batch-motion-model-v1-protocol.md) has a final
[machine-readable result](mmm-batch-motion-model-v1-final.json): **13/13
finite controls pass**. This advances the Goal 6 batch-authority design;
it does not change runtime code or establish loaded performance.

The model assigns one consecutive runtime/evidence version, motion time,
and source-touch version to each newly accepted event in a same-tenant,
same-time batch. Exact duplicates use no version. It tests mixed
duplicate/new input, changed-digest and malformed-input rejection,
future-only versus visible-backfill as-of reads, and protected-read
exclusion while the guard is held. A protected read after release accepts
an older capture only when **every** intervening version has future-only
motion. The [Go bridge test](../../internal/store/research_batch_motion_contract_test.go)
confirms this per-version condition against the actual
`store.ResearchSnapshotCompatible` function.

The proposed model order is backend commit, durable lineage commit,
in-memory publication, then acknowledgment. After a crash before backend
commit, reopening retains the old checkpoint. After backend commit but
before durable lineage, reopening fails closed because the checkpoints
disagree. After durable lineage but before publication, reopening can
reconstruct complete motion from the sidecar and resolve an exact retry.
A possibly committed backend error cannot clean-abort or acknowledge.
Duplicate-only retries instead prove no version motion and clean-abort;
they never perform a no-op durable or publication commit.

## Corrections retained

The first [artifact](mmm-batch-motion-model-v1.json) is superseded: its
test helper ran duplicate-only retries through no-op durable/publication
steps, contrary to the frozen protocol. The intermediate
[artifact](mmm-batch-motion-model-v1-corrected.json) fixed that path but
did not stage the entire durable map before validating later entries.
The final source stages backend and durable changes and includes a forced
second-entry failure that leaves durable motion and touches unchanged.
The two earlier artifacts are retained as audit history but do not replay
with the final source. Only the final artifact is the result cited here.

## Scope and next step

Backend and lineage commits are atomic *model steps*. The experiment does
not prove their real database transactions, power-loss recovery, cross-DB
atomicity, async scheduling, unguarded compatibility callers, or exact
`Service.Observe` transformation. A backend/sidecar checkpoint mismatch
remains quarantined and needs an external reconciliation procedure; the
model does not manufacture one. The next isolated implementation should
exercise batch-aware publisher and lineage transactions with real SQLite
and raw LibraVDB, including uncertain commit and reopen, before any shared
runtime API or 4 ms loaded screen is changed. Production is untouched.

Verification completed:

```sh
node research/batch-motion-model-v1.mjs --replay docs/experiments/mmm-batch-motion-model-v1-final.json
go test ./internal/store ./internal/researchpublication ./internal/researchlineage -count=1 -timeout 3m
go vet ./internal/store ./internal/researchpublication ./internal/researchlineage
```

Final artifact SHA-256:
`a60d0a8456fd4427d639754417ded3ce14e2f3e3b8ab2745dd99bf2f3d9a957b`.
Final model source SHA-256:
`e02c796fb483423454d644fb053c7769cf4ddf2c6e48112a72bfef32efe8b5d8`.
