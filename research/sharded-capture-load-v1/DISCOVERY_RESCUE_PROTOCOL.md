# Logical parent discovery rescue

Frozen before regression execution, 2026-10-05. All work is in the temporary
v1.6.13 dependency copy, never the installed module cache or production.

## Patch Reasoning Gate

Confirmed: the original topology test cannot reopen its tenant collection.
`ListCollectionsWithContext` removes all physical shard names without adding the
logical parent when that parent has not yet been loaded. The wrapper concludes
the collection is absent and tries to recreate existing children. Direct
`GetCollection(parent)` can reopen these children in the admission regression.

Alternative causes were wrong tenant naming, incomplete children, and lost
quantization metadata. The first two do not explain successful direct parent
reopen on the same routing mechanism; quantization persistence remains a
separate, unproved follow-on issue. The exact pinned listing function also
matches upstream commit 18515d4ce0620b24fd2512b0f9023b8b1f4d8355; see the
read-only source comparison. Earlier all-state issue/PR inspection is recorded
in UPSTREAM_CHECK.md, not an exhaustive claim about unpublished fixes.

Invariant: collection discovery returns each logical parent once, never its
hidden physical children. Discovery does not create records or hide load
errors. Existing nonsharded names, sorting, cancellation and closed checks stay.
Deduplication remains O(number of physical collection names), off the warmed
request path. No migration, shard deletion or production feature enablement.

## Falsifiers And Gates

The new library regression must fail before the repair and pass afterward:
create one ordinary and one four-shard collection; insert public probe records;
close and reopen without loading parents; assert both sorted logical names,
complete iteration, explicit parent loading, cancelled discovery and closed
database errors. No skipped suite counts as passing. Run it with race detection.

Then run the ORIGINAL strengthened eventframed topology/reopen test against the
repaired temporary dependency. Its SQ8, idempotence, future exclusion, tenant,
snapshot and persisted deletion assertions are unchanged. Any failure blocks
promotion even if fresh-load timings improve. Preserve original timeout and
reopen failures. A positive component result is not whole-goal validation.
