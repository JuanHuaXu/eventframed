# Batch intent prototype v1: frozen storage protocol

This Goal 6 experiment adds a test-only Go prototype in its own package. It
must not change `Service.Observe`, the production backend constructor,
`researchpublication.Publisher`, or `researchlineage.Ledger`.

Use real LibraVDB `PutResearchEventBatch` and a separate private SQLite
WAL/FULL sidecar. A caller owns both stores exclusively. For a same-tenant,
same-availability-time batch of 1..16 ordinary events, validate request
identity and exact duplicate payloads, then durably record one intent with
the before snapshot, expected after snapshot, ordered new records and full
input. Commit the backend batch; verify its duplicate mask, final snapshot,
and exact accepted payload/vector readback. In one SQLite transaction,
record every accepted version's motion and source identity, advance the
sidecar checkpoint, and remove the intent. Acknowledge only afterward.

On reopen, a pending intent may be removed only if the backend is still at
the before snapshot and none of its proposed new IDs exists. If the backend
is at exactly the expected after snapshot and every accepted record matches,
complete the sidecar transaction. Any other state fails closed. A sidecar
without an intent must exactly match the backend checkpoint; missing sidecar
is never silently recreated. Duplicate-only retries consume no version or
intent. No intermediate version may be omitted from as-of motion.

Test crashes after durable intent, backend commit, and sidecar finalization;
exact retry after each recoverable crash; changed-digest and changed-payload
conflicts; mixed duplicate/new input; future-only and visible-backfill
as-of checks; altered pending intent; and backend/sidecar mismatch. Assert
raw backend snapshot, accepted IDs and vectors after reopen, sidecar motion,
and zero false acknowledgements. Run focused race, ordinary package tests,
vet, and an isolated per-batch timing comparison against raw batch writes.

The prototype may assume one process and an exclusive temporary directory.
It does not prove cross-process ownership, hardware power-loss behavior,
arbitrary mixed availability times, `Service.Observe` conversion parity,
loaded writer tails, or 4 ms learning freshness. A pass is not authorization
to install it in shared runtime code.
