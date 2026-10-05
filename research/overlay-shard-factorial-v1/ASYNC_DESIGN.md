# Next Isolated Background-Merge Invariants (Not Implemented)

The frozen factorial still uses synchronous V6 compaction. Do not present this
design as measured code or a rescue of a failed outcome. Use the observed
interaction results first; retain their unchanged gates and negative controls.

## State And Publication

A prospective generation has immutable base at committed watermark a, frozen
read-only delta covering committed writes through b, and bounded mutable-tail
view covering every later committed write through c. Entries carry mutation
sequence/identity, canonical storage ordinal and immutable owned vector. Latest
committed mutation wins across all three layers; a tombstone shadows older IDs
and storage ordinals. Index ordinal renumbering is translated at every boundary.

Only a successful WAL/transaction commit publishes a tail mutation. A merge
constructs a replacement base from the committed b snapshot, NEVER prepared
uncommitted rows or a live provider. At publication reserve the same writer
boundary as Prepare/Commit/Abort, validate the captured generation fence, then
publish new base plus ALL tail mutations (b,c]. A maintenance swap cannot change
which transactions were acknowledged or advance an evidence/causal epoch by
itself. If the fence is stale, discard/retry; no partial rebasing or dropped IDs.

Separate monotonic per-ID stamps are required: pointer comparison alone cannot
reliably distinguish repeated nil tombstones or ordinal reincarnation. Mutation
ordering must include delete/upsert/delete, repeated IDs and ordinal replacement.
All canonical public scores are hydrated through the original Collection path;
physical graph IDs must never be exposed as external storage ordinals.

## Bounds And Lifecycle

Use ONE bounded compactor initially, not four unbounded route goroutines. Declare
separate frozen and tail caps and their combined maximum. If tail is full,
explicit context-bounded backpressure or transaction error is required before
WAL acknowledgement; never silently truncate or drop evidence. Every rejected
write and queue wait belongs in throughput/deadline counts, not just successes.
Merge rate must outrun sustained admitted mutation rate; otherwise this design
is not a sustained-freshness solution regardless of cheap pointer publication.

Reader leases retain old base until its last user leaves. Charge current,
under-construction, frozen/tail and retired generations to byte/RSS accounting.
Cancellation stops construction; Close cancels and joins the worker without
holding locks needed by that worker. Retired close errors do not retroactively
erase a committed transaction. Snapshot/reopen must recover ALL durable data,
including a merge interrupted before/after publication; ANN topology persistence
is not promised unless separately defined and tested.

## Required Falsifiers

Deterministic blocked-build hooks: admit a new upsert/delete while merge is held;
verify reads before/after publication and reopen, aborted WAL stays invisible,
no committed update disappears. Full-tail pressure must produce the declared
bounded response; cancellation/Close must terminate with readers in flight.
Sparse external ordinals and future-prefix sorted compaction must preserve native
filter/hydration identity. Run ordinary/race, real mixed transaction, durable
ledger, chronological quality and full loaded integration before measuring wins.
Then sustained open-loop load with offered/completed/rejected/deadline counts,
compaction intervals, heap/RSS, bytes and abrupt-process recovery. No threshold
relaxation, future filtering shortcut or uncharged background acquisition cost.

Primary design inspiration: FreshDiskANN sections5.1-5.3,
https://arxiv.org/html/2105.09613v1. Its immutable long-term and RO/RW temporary
indices support the architecture; its measured scale and latency are not
guarantees for this research HNSW wrapper or EventFrame's learned predictions.
