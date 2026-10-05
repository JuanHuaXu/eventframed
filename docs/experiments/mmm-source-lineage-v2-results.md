# Durable research source lineage v2: finite result

2026-10-01. Frozen protocol:
[mmm-source-lineage-v2-contract.md](mmm-source-lineage-v2-contract.md).
The opt-in research sidecar PASSES the frozen local restart and fail-closed
screens. It does not complete Goal 6 or authorize production adoption.

The sidecar uses an absolute private local SQLite file with one process owner,
FULL synchronization, a committed backend snapshot, and per-event last-touch
versions. The wrapper appends the checkpoint and touch set in one SQLite
transaction after each successful backend version change, under its existing
writer gate. The default wrapper constructor and production path are
unchanged. Sidecar mode does not duplicate event touches in RAM.

**Correctness evidence.** On temporary persistent LibraVDB, the wrapper
reopened with the same sidecar after A/B writes, a non-event policy version
change, deletion of A, and deletion/recreation of B. Unrelated B remained
continuous after A was deleted; B's original identity did not reappear after
same-ID recreation. A service-level integration test admitted two witnessed
forecasts, processed two guarded delayed labels, reopened the SQLite label
stream and source backend, and rebuilt a new epoch from B only after A's
deletion. A later restart after B's exact recreation retained zero old labels.

An injected sidecar failure **after** a backend commit faulted the current
wrapper; reopening against that backend rejected the checkpoint mismatch.
Second-owner lock conflict, missing/corrupt checkpoint, stale predecessor,
and stale target controls also rejected. Duplicate `Put` still quarantines
publication without recording a touch, preserving the v1 negative result
rather than silently treating an unresolved duplicate as a harmless no-op.

**Cost screen.** Three ordinary-build repetitions of 200 operations on an
Apple M4, with an in-memory event backend, measured:

| Operation | In-memory wrapper | SQLite sidecar wrapper |
| --- | ---: | ---: |
| Positive continuity check | 69.58-73.75 ns/op | 4.741-5.456 us/op |
| Distinct event `Put` | 2.587-3.400 us/op | 61.973-73.008 us/op |

These are uncontended mean benchmark times per operation, not p95/p99 or a
loaded service latency distribution. The sidecar is off the ordinary default
path. Focused race tests, vet, and full `researchlineage`,
`researchpublicationstore`, and `service` package suites passed after the
final memory-map removal.

**Limits.** This is a two-database ordered protocol, not an atomic transaction
across LibraVDB and SQLite. The injected failure demonstrates fail-closed
behavior for the tested commit gap; it is not a power-loss/fsync simulation.
Positive continuity assumes exclusive backend ownership and no cloned backend
with recycled version history. The sidecar stores tenant/event IDs in a private
local file; it is not a PII-free public artifact. Key rotation, missing
sidecar recovery, crash testing, guarded model installation, loaded p99, and
untouched agent-task outcomes remain open. At the v2 checkpoint, the old
publisher's as-of motion history was not restored; the separately tested
[v3 extension](mmm-source-motion-v3-results.md) addresses bounded
future-ingestion motion on the opt-in sidecar.
