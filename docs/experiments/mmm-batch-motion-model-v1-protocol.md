# Batch motion authority model v1: frozen protocol

This is an isolated Goal 6 state-machine experiment. It does not modify
the runtime publisher, durable lineage ledger, backend, or service. The
purpose is to derive a batch contract that retains the existing per-version
as-of and source-identity invariants before a Go implementation is attempted.

Model one owned, same-tenant batch of 1..16 ordinary events at a shared
nonzero availability time. The backend transaction either inserts all new
events or none; exact-digest duplicates consume no version, while an ID with
a changed digest rejects the entire batch. Accepted new events get versions
`v+1,...,v+m` in input order and each gets its own motion and touch record.
Runtime version and evidence epoch must move together by `m`. No general,
graph, policy, or posterior mutation may be hidden inside this operation.

Model the phases `begin -> backend commit -> durable lineage commit ->
in-memory publish -> ack`, with an owned guard held through the operation.
Protected as-of admission is unavailable while the guard is held. After
release, a captured snapshot is compatible only if every intervening
version has recorded availability strictly after the query's `asOf` and
the semantic versions match. A duplicate-only batch proves no backend
motion and aborts without advancing any version.

Enumerate crash/reopen after each phase. Reopen succeeds only when durable
lineage and backend snapshots agree; before durable lineage catches a
committed backend, it must fail closed pending external reconciliation.
A crash after durable lineage but before in-memory publish may reconstruct
the publisher from the complete durable motion. A possibly committed
backend error must never become a clean abort or success acknowledgment.

Required controls: exact retry, changed-digest conflict, intra-batch
duplicate, mixed existing/new entries, unequal availability times,
oversized batch, zero-time input, incomplete motion, final-version-only
touch mapping, visible backfill, and an attempted protected read at every
phase. The source hash, protocol hash, and full deterministic result must
be preserved. Passing proves only the stated finite model and its asserted
invariants, not SQLite atomicity, LibraVDB crash behavior, service conversion,
concurrency scheduling, or loaded latency.
