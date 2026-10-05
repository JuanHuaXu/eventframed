# Last-published-LSN read rescue v1: frozen private protocol

The [open-loop v1 screen](mmm-sort-batch-openloop-v1-protocol.md) failed its
read-tail ratio twice. A separate phase profile attributed nearly all read
time to `capture`, not the pinned exact-LSN SQL. Test one distinct
research-only reader: after full journal/marker publication, atomically
expose an immutable `(LSN, published_at)` pair. A read may pin that last
certified LSN without waiting on the writer gate. It must reject absent or
older-than-250-ms publication state. A writer error poisons the in-memory
slot; restart still requires full journal verification. The slot never
publishes a DB-only commit or a partially updated SQLite marker.

Functional controls: initial verified view; subsecond exact-LSN as-of;
read of the previous certified LSN while a later DB commit is pending;
poison after DB-only error; expiry after 250 ms; close/reopen verification
of a fully committed batch. A same-process out-of-band Store write is
outside the slot's writer API and must not be treated as authenticated
freshness; this is not an external-writer defense.

Then run three fresh matched 4 ms offer blocks, rotating three arms:
one-event journal plus original gate read, cap-16/16-ms-dwell journal plus
original gate read, and the same batch journal plus last-published-LSN read.
Keep 200 eligible base rows, four future sentinels, 256 visible writes,
192 searches, eight read workers and all open-loop v1 ID/row/reopen checks.
The slot arm acknowledges a group only after the published pointer is
updated. Report actual offer gaps, writer offer-to-ack p50/p95/p99, search
call and offer-to-completion p50/p95/p99, observed published-view age
p50/p95/p99/max, all integrity violations and per-arm search errors.

The rescue component passes only if it has zero integrity violations and
search errors, all 256 acknowledgements, candidate writer age p99 below
250 ms, search-call p99 below 100 ms, published-view age max below 250 ms,
and search-call p99 no more than 1.10 times the single-writer original-read
control. Report the unchanged batch/original-read arm even if it fails.
These are private backend gates, not full Recall or learning-freshness
criteria. The one-owner and no-raw-bypass boundary remains; production is
untouched. Preserve failure and avoid threshold retuning on these blocks.
