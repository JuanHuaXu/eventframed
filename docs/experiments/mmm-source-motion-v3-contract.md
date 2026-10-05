# Durable future-ingestion motion v3: frozen research contract

2026-10-01. Goal 6 continuation. V2 preserves event identity across an
orderly wrapper restart but does not restore the publication proof that every
intervening runtime version was an ingestion available strictly after a
forecast's `as_of`. Delayed feedback from a valid old forecast therefore
fails even when the intervening writes were future-only.

Only the opt-in durable research wrapper may restore motion. In the same
SQLite transaction as the v2 checkpoint and event touches, record the
availability time for a successful `Ingestion` commit at its exact runtime
version. General mutations have no motion row. Keep at most the publisher's
existing 4096-version window; do not invent older history. On open, first
match the backend snapshot to the sidecar checkpoint, then load only valid
motion rows into the publisher. Sidecars created before this extension remain
usable for v2 source continuity but restore no motion. The default wrapper
must retain its existing fail-closed restart behavior.

The model invariant is unchanged: an old snapshot is as-of compatible only
when **every** intervening version is an ingestion with recorded
`AvailableAt > as_of`, the evidence-epoch delta matches the runtime-version
delta, and all semantic versions match. Any general mutation, missing row,
zero time, stale sidecar, or version gap fails closed. The source witness and
event lineage checks remain independent requirements for a transferred
label. A true as-of proof is not permission to publish a learned model.

Frozen checks before results:

1. A persistent A/B forecast admitted before a future-only C ingestion can
   receive delayed feedback after orderly backend/wrapper restart. Verify the
   original journal and source witness at the service boundary, then replay
   the label durably. The same sequence without a sidecar must still reject
   the old forecast.
2. An ingestion at or before `as_of`, a general version change, and a missing
   or corrupted motion row each reject old-snapshot authority after restart.
   A sidecar created by v2 without the motion table cannot gain retrospective
   authority. Current-snapshot authority remains available.
3. The sidecar write-gap failure from v2 remains fail-closed. Focused race and
   vet checks, full affected-package tests, and separate write/validation
   cost measurements are required. No production path, real user data, or
   OpenClaw instance is used.

This experiment tests orderly restart and injected gaps, not power loss,
multi-process backend ownership, population p99, or a full learner handoff.
