# Bound worker v7: motion-proof diagnostic

Date: 2026-10-01. Protocol: [v7 pre-run diagnostic](mmm-bound-worker-motion-v7-protocol.md).
No production or serving change.

## Result

The memory and persistent-store runs agreed:

| Transition from original forecast | Source witness and event continuity | Publisher as-of compatibility |
|---|---:|---:|
| No mutation | valid | compatible |
| New event available one hour later | valid | compatible |
| Backfilled event available before forecast | **still valid** | **incompatible** |

The backfill leaves the witnessed source event unchanged, so
`ValidateResearchTransferSource` correctly answers its narrow source question.
It does **not** certify that the original query's observation set is still
complete. The independent publication motion map correctly rejects the
backfill. This confirms that simply removing the target snapshot from the
v5/v6 bootstrap HMAC would let a backfilled history reuse the old bound model.
That rescue is rejected.

```text
go test ./internal/service -run '^TestResearchBoundWorkerMotionProofDiagnostic$' -count=1 -v
```

Both subtests passed and logged `source_valid=true asof_compatible=false`
after backfill. The result is a component counterexample, not an answer-quality
or latency measurement.

## Required v7 design

Cross-snapshot continuation must persist the bootstrap's **origin snapshot**
separately from a target-independent evidence commitment. At every reopen,
the service must prove bounded as-of compatibility from origin to current
target at the bootstrap cutoff, validate every retained source against the
current target, and check each new-epoch original against its own admission
snapshot and prediction time before replay. A changed retained decision,
missing motion history, backfill, general mutation or witness mismatch must
fail closed. The final target and source-log sequence still need an atomic
store-before-source certification before exposing the worker.

The current ledger stores only the seal, not the origin snapshot. Old research
logs cannot be assumed to carry that evidence; migration needs an exact-target
proof or must explicitly reject cross-target reopening. Until this is built
and tested, v5/v6's exact-target seal remains in place. Goal 6 remains open.
