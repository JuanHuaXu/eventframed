# Durable future-ingestion motion v3: finite result

2026-10-01. Frozen protocol:
[mmm-source-motion-v3-contract.md](mmm-source-motion-v3-contract.md).
The opt-in research sidecar PASSES the local orderly-restart and fail-closed
screens for future-only ingestion motion. Goal 6 remains open.

The sidecar now records an ingestion's exact runtime version and
`AvailableAt` in the same SQLite transaction as its source-touch set and
checkpoint. A durable wrapper reopens only after matching the backend
snapshot, then builds a publisher from the recorded, bounded motion map. The
default wrapper continues to start with no pre-restart motion. An older v2
sidecar gains an empty motion table, never retrospective proof for earlier
writes.

**Observed behavior.** In temporary persistent LibraVDB, an old snapshot
remained as-of compatible after future-only ingestion and wrapper restart.
An ingestion available at or before the forecast's `as_of` rejected that old
snapshot, before and after restart. A general policy-version change, a
missing motion row, and an older sidecar without the table also rejected old
authority while permitting current-snapshot reads. A malformed motion row
prevented durable-wrapper startup. Publisher construction rejected invalid
version/time entries and detached its map from caller mutation.

The service integration crossed the more important boundary: an original
forecast and keyed source witness were admitted before a future-dated event
was inserted; after orderly backend/wrapper and SQLite-label-stream reopen,
the original journal/source validated and one delayed label was accepted and
processed under the as-of guard. Reopening the same backend through the
default research wrapper still rejected that old forecast. These are
synthetic fixture labels, not real agent-task outcomes. The v2
post-backend-write failure and same-ID resurrection controls continued to
pass in the affected full suites.

**Finite component cost, Apple M4, ordinary build, three repetitions of 200
operations per cell:**

| Operation | Observed mean range |
| --- | ---: |
| Sidecar event `Put`, including motion row | 76.8-89.0 us/op |
| In-memory event `Put` control | 2.45-3.18 us/op |
| Sidecar source-continuity check | 5.45-6.57 us/op |
| Restored as-of check across 1 version | 31-33 ns/op |
| Restored as-of check across 128 versions | 0.91-1.06 us/op |
| Restored as-of check across 4096 versions | 29.5-38.5 us/op |

Focused race tests, vet, and full `researchlineage`, `researchpublication`,
`researchpublicationstore`, and `service` package suites passed. The benchmark
is uncontended component work, not end-to-end Recall p99. The 4096-version
bound preserves the prior publisher's behavior: an older snapshot outside the
retained history fails closed.

**Still open:** actual power-loss crash testing, cloned or out-of-band backend
histories, key rotation, bounded loaded latency/freshness, publication of a
transferred model under the final source guard, and untouched outcome-labeled
agent tasks. The sidecar and backend are two separately committed databases;
matching snapshots detect the tested missing-sidecar-commit state but do not
make the pair atomically durable under arbitrary storage failure.
