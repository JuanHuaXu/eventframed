# Post-failure admission rescue

Frozen after the original candidate preflight, before rescue execution.
Original store test timed out in TestResearchEventBatchParityReopen. The Go
stack shows commitTxWithGraphReceipt acquiring multiple shard-view permits
from the same parent write controller (capacity two), holding earlier permits
while awaiting the next. Control passes. This is a confirmed self-blocking
transition, not slow HNSW computation or corruption of an existing database.

Invariant: one atomic transaction acquires one admission permit per DISTINCT
write controller, while still locking every physical target and preserving
PrepareTx, WAL commit, index publication, cancellation and release. Shard views
share their parent's immutable controller pointer. Do not merely increase the
pool capacity or delete per-collection mutation locks. The fix is tested in a
TEMPORARY dependency copy; installed module cache and production stay unchanged.

First add an explicit regression to the unmodified temporary v1.6.13 source:
one-slot controller, four declared shard routes, bounded 200 ms transaction.
It must fail before the patch by deadline without committing partial records.
After pointer-deduplicated admission, it must complete, preserve all four
records, release permits, survive reopen, and reject a cancelled mutation.
Run ordinary and race builds; keep the original failing log. No timeouts are
relaxed to turn the failure into a pass.

Then use a separate eventframed candidate checkout with the original sharding
patch and a local module replacement. Original candidate sources/results remain
immutable. Run full store regression again with its original six-minute limit
and separately retain topology/reopen outcome. Any remaining reopen/quantization
failure blocks promotion regardless of fresh-load performance. The original
benchmark runner is not executed if its source-hash contract no longer matches.

Separate diagnostic already found the candidate cannot reconstruct its event
collection after reopen: it attempts creation and finds existing shard storage.
That is a second confirmed integration issue, NOT yet proof of quantization
loss (the earlier error occurs before that check). No broad migration or durable
format change is authorized by this pilot. Investigate those objects separately.
