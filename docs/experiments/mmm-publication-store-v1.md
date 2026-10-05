# Core-method research publication adapter

`internal/researchpublicationstore` is a new, unconfigured research adapter. No
daemon constructor installs it; production stores and plugin contracts remain
unchanged. It is separate from the frozen ingestion-only load adapter.

## Coverage and behavior

All31 core EventStore methods are classified by a reflective interface test.
Sixteen general mutation/maintenance methods and ordinary Put have explicit
publication wrappers. Close has a quarantine/resource-release path. Journal
writes retain their real backend validation/persistence but do not invalidate
forecast-state versions. Twelve read methods delegate without new authority.

An AST coverage test requires the expected General/Ingestion guard on each
wrapped mutation. A newly added core interface method or missing override fails
the test instead of silently inheriting an unsafe write path.

The adapter must exclusively own its backend; direct writes through a retained
backend handle bypass publication and are outside the contract. General changes
block compatibility while pending. Only ordinary ingestion receives the strictly
future-event exemption. Composition is general, despite adding a new event.

Errors and panics quarantine rather than assume rollback. A successful general
operation with unchanged authoritative version can finish as a no-op. An ingestion
with unchanged version (including a duplicate) is conservatively unresolved and
quarantines. That is an availability limitation requiring explicit recovery,
not production-compatible duplicate handling. Close still releases resources
after quarantine. There is no automatic optimistic reset.

## Evidence

Three race-test repetitions passed, plus vet:

- Interface classification and source-level guard checks.
- Backend panic injection through every wrapped writer and Close, verifying that
  no compatible publication survives the uncertain outcome.
- Actual memory and persistent store policy binding, future ingestion, deletion,
  repeat ingestion and duplicate-quarantine paths.
- Version invalidation after policy/deletion and preserved earlier-as-of proof
  for future-only ingestion.

Successful-path parity tests additionally compare wrapped and plain memory and
persistent stores for policy binding/no-ops, selection/omission/Anti-Pigeon
certificates, composition, decomposition and compaction. Returned values and
snapshots match; changed versions invalidate earlier proofs in these fixtures.
These tests passed three race repetitions. Additional posterior/snap fixtures
now compare outcome application, duplicate application, posterior read, graph
publication, residual invalidation and rollback. Graph IDs, node count, version
increments and dependent residual inactivity are explicitly checked. The
post-snap posterior is uncertified; this fixture does not prove a transition
from an initially certified posterior, so certified-state coverage remains open.

The first rollback parity run failed because separate backend calls generated
different wall-clock PublishedAt values. Source inspection confirmed both
backends assign time.Now().UTC(). The corrected test validates each timestamp
against its call interval and exact graph readback, then normalizes only that
field for cross-store comparison. No backend behavior was changed. The corrected
suite passed three race repetitions and vet. Restart/crash coverage and the full
mutation boundary matrix remain incomplete.

Agency success-path parity now covers signed fixture insertion, duplicate
insertion, claim, leased no-op, approval and duplicate resolution on both real
backends. Explicit assertions require pending/claimed/approved statuses and the
expected duplicate flags, rather than relying solely on matching outputs.
The existing parity harness checks snapshots, current proof validity and old
proof invalidation on version changes. Three race repetitions and vet passed.
No agent or executor is connected. These checks do not establish restart/crash
recovery of the adapter, lease expiry/reclaim behavior or autonomous safety.

The capability-aware `Wrap` factory preserves vector retrieval only when the
backend implements it. Tests compare returned vectors with each real backend and
verify that a core-only backend does not acquire a fictitious vector capability.
The older research snapshot-compatibility hook delegates to the same publication
proof. Three race repetitions and package vet passed after this addition.
This tests vector forwarding, not every possible optional extension.

## Remaining gaps

No mixed-mutation load benchmark, complete successful-path suite, atomic crash
recovery or persistent replay ledger has run against this adapter. Optional
capabilities beyond the tested vector and research-compatibility hooks still need
an explicit forwarding/availability audit. The prior queue16 load pass belongs to
the ingestion-only adapter and must not be transferred automatically to this one.

This closes a core method-interception gap in research code; it does not complete
roadmap6 or authorize deployment, real-task claims, or a GitHub push.
