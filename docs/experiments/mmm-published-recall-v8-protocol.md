# Published-LSN Service Recall v8: frozen deterministic contract

Date: 2026-10-02. Test-only adapter and isolated service fixture; no
production/default runtime change. Prior ID-only and payload SQL results
do not prove that `Service.Recall` binds candidates, runtime snapshot,
and durable Bayesian frontier journal to one state.

## Declared research boundary

Use the private 256D declared-payload, unit-normalized, journaled gate.
A single owner serializes every event append, frontier-journal append,
and publication-pointer update. Each published pointer contains its
certified LSN, `model.Snapshot`, and publication time. Each Recall call
gets its own mutable context pin. Search captures one pointer, queries
exactly its LSN and stores that pointer in the call's pin. Snapshot
returns the pinned snapshot for that call. Frontier journals go through
the authorized metadata-only gate and republish the pointer. Before
entering the gate, check `JournalSnapshotCompatible` against the current
runtime state under the same owner lock; a stale read returns
`ErrStaleSnapshot` without poisoning a gate that has not moved. No raw
Store writer is permitted in this fixture. Graph, certificate, posterior
and residual states remain constant during the two interleavings.

## Fixed controls

Use `Service.Recall` with an explicit 256D query embedding, three
initial rows (two eligible, one future), recall/pack caps three, and
an as-of at 120 ms after 2026-10-02T00:00:00Z. A test hook runs after
the first SQL Search but before `Recall` reads its snapshot.

1. Visible interleave: append one as-of-visible row and republish.
   The first journal attempt must reject the old pin as stale, the
   service must retry, and the successful packet and durable journal
   must both use the new snapshot and include the visible row in the
   nominated frontier. The gate must remain READY.
2. Future-only interleave: append one row available 5 ms after as-of
   and republish. The first attempt may complete on the old pin without
   a retry, but neither the packet nor journal may include the future
   row. The packet/journal snapshots must agree, and the gate must
   remain READY. A later recall on the new pointer must also exclude
   the future row at the same as-of.

Pass only if both controls have zero errors, exact expected search and
stale-retry counts, matching packet/journal snapshots, expected frontier
membership and no future leakage, and an authoritatively capturable
READY gate after each journal. Record focused normal/race, ordinary
package tests and vet. Do not retune the hook timing, caps or vectors.

This is a deterministic event-only service integration, not loaded
multiworker performance, graph/posterior mutation isolation, full agent
retrieval, or Goal 6 completion. A pass still requires a broader read-
view API and writer ownership mechanism before any production proposal.
