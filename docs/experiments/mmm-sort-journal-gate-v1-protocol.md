# Journal-through-gate publication v1: frozen private protocol

The direct `PutBayesianJournal` coexistence probe failed: a valid metadata
write advanced LibraVDB's global LSN without changing EventFrame runtime
versions, permanently staling the event-only marker. Test a distinct
test-only transition under one exclusive gate owner. Before a frontier
journal write, verify the current READY marker, paired Store snapshot and
global LSN. Write through the real `Store.PutBayesianJournal`, read back the
exact journal content, and pair the post-write snapshot/LSN. The snapshot
must be unchanged. Atomically advance only the sidecar marker LSN; leave
the EventFrame row count and hash-chain root unchanged. An exact duplicate
journal must neither commit nor move the marker.

Functional controls: new journal followed by a normal EventFrame batch and
close/reopen verification; exact duplicate; conflicting duplicate; direct
journal bypass; injected interruption after DB journal commit and after
SQLite marker commit. DB-only uncertainty must fail closed, and SQLite
completion may recover only after full journal verification on reopen.
An out-of-band raw DB mutation or second independent Store is outside this
single-owner contract and must not be silently treated as an authorized
journal.

Measure 100 isolated new-journal gate calls after 16 warm-ups, reporting
p50/p99 and quiet published-view read cost. This is a research component,
not full Service Recall, loaded performance, cross-database atomicity,
multi-owner fencing, or production deployment. Preserve the direct-path
failure and any new negative results.
