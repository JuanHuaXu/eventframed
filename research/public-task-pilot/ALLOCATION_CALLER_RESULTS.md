# Allocation caller audit

Evidence: `allocation-caller-analysis/alloc-callers.txt` plus prior raw profiles.
Fresh public upstream metadata is retained alongside it. HEAD remains
18515d4ce0620b24fd2512b0f9023b8b1f4d8355; seven PRs and ten issues including PRs,
with no pagination link. No separately titled fix for these allocation paths was
found; this is not an exhaustive semantic review of every upstream change.

## Confirmed Paths

The6400 diagnostic differential allocation profile attributes about0.58GiB
(36.67% of its sampled total) cumulatively to HNSWBase.Close -> Database.Close ->
singlefile.Engine.Close -> checkpointLocked. Do not sum cumulative stack entries.
Engine.Close checkpoints dirty/catalog-dirty state for close/reopen, collecting
committed graph WAL and serializing snapshots. Relaxing WAL sync does not remove
this work; Close also contains a final dirty-file Sync path.

PartitionLease.Release calls closeRetired synchronously when it releases the last
reader of an old graph. Therefore checkpoint/serialization may execute on a read
request's completion path. This is a confirmed call-path exposure, NOT proof that
it caused any particular earlier100ms outlier. Compact can also close a graph.

About0.50GiB (31.89%) sits under PartitionLease.Search, including0.42GiB through
storage.Collection.Get and cloneEntry. Public collection search hydrates returned
vectors/metadata; the research adapter then ignores those payloads and rescales
scores from its owned immutable map. The extra payload is redundant for THIS
adapter, not for general mutable collection callers. Bypassing it globally would
violate the public search contract.

## Rejected Shortcut

Record-put encoding already estimates vector bytes and metadata before acquiring
its buffer. WriteUint32's allocation attribution alone does not prove a missing
preallocation bug. Do not replace the encoder or weaken ownership copies on this
evidence. Snapshot estimates and detached-buffer lifetime need separate tests if
that path is pursued.

## Next Narrow Experiments

1. A bounded retirement worker can move close/checkpoint off lease release while
   keeping pending/closing graphs charged. It must join on shutdown and expose
   errors; moving work is not removing its CPU/memory cost or solving throughput.
2. A candidate-only interface for the private immutable derived HNSW base could
   omit payload hydration while retaining normalization, ID ownership, nomination
   and error semantics. Existing public collection search must remain unchanged.
3. Skipping checkpoint would require an explicit disposable-database lifecycle,
   not treating every no-sync database as disposable. No such change is made.

No dependency/module-cache or production patch was made. The available evidence
supports these bounded experiments, not a universal backend fix. All seven whole
research goals remain open.
