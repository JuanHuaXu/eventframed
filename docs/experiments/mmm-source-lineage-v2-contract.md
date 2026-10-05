# Durable research source lineage v2: frozen experiment contract

2026-10-01. Goal 6 continuation. V1's in-memory mutation ledger rejects
same-ID resurrection in one wrapper lifetime but cannot certify an original
after wrapper restart. This experiment is opt-in and research-only. It does
not alter a production store schema, daemon constructor, or OpenClaw path.

The wrapper exclusively owns a backend and an absolute-path SQLite sidecar.
One process owns the sidecar at a time. The sidecar stores an initial backend
snapshot, the last fully recorded backend snapshot, and the latest successful
touch version for each `(tenant, event ID)`. Every successful version-changing
backend mutation is followed, under the same writer gate, by one SQLite
transaction updating the checkpoint and all touched events. This includes
non-event version changes with an empty touch set. A new wrapper may resume
only if its backend snapshot exactly equals the sidecar checkpoint. A mismatch
is a terminal, fail-closed result, not permission to silently rebase the
sidecar. If the sidecar write fails after a backend commit, the current wrapper
also fails closed for source continuity.

This is not a cross-database atomic transaction. It relies on append ordering:
backend commit first, sidecar commit second. A crash in the gap must leave a
backend/sidecar snapshot mismatch, hence no positive continuity proof. A
positive result also still needs the original keyed source witness, the
original journal, exact current event, and an installation-time publication
guard. Matching snapshots cannot detect a cloned/replaced backend with a
reused version history; exclusive backend ownership and stable database
identity are explicit experimental assumptions.

Frozen checks before results:

1. Restart a persistent backend and wrapper using the same sidecar after
   ordinary A/B admission and labels. Delete A, restart again, and rebuild
   from B only. Delete and recreate B with the same ID/fields; B must remain
   revoked across another restart. A non-event version change must not revoke
   B but must advance the sidecar checkpoint.
2. Inject a sidecar-write failure after a successful backend commit. Current
   positive checks must stop. Reopening against that backend and old sidecar
   must reject the snapshot mismatch. A sidecar lock conflict and corrupted
   checkpoint must also fail closed.
3. Preserve v1's negative duplicate result: a duplicate `Put` may quarantine
   and is not promoted to a benign no-op by this experiment. No production
   mutation semantics are changed.
4. Run focused race, vet, and affected-package suites. Measure validation
   and mutation overhead separately from service tail latency. A local
   component pass is not Goal 6 completion.

SQLite transaction durability depends on its sync and filesystem settings;
the experiment uses a local file with FULL synchronization and does not claim
network-filesystem safety. Relevant primary references:
[SQLite atomic commit](https://www.sqlite.org/atomiccommit.html),
[SQLite WAL and synchronous modes](https://www.sqlite.org/wal.html), and
[SQLite corruption caveats](https://www.sqlite.org/howtocorrupt.html).
