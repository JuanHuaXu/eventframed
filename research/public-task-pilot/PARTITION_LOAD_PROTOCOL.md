# Partitioned compaction-inclusive growth screen

Reuse GROWTH_PROTOCOL.md unchanged: initial800/3200/6400, two repeats,768d,
CPU4,1024 reads every5ms,512 writes every10ms,100ms due-relative deadlines,
delta64 TOTAL, trigger32 TOTAL, one builder, two retired graphs, eight leases.
No retries. Eight fixed ID partitions, all queried. Same durable flat+metadata
transaction callback, bounded setup transactions and authoritative reopen audit.

At each5ms compactor poll, if total delta>=32, select the partition with the
largest delta (lowest index breaks ties). Build only it. Log total and selected
delta before/after, global revisions and elapsed time. Concurrent writes mean
before-minus-after is a net change, not the number of entries compacted.
Record per-arm sidecars immediately. No other workloads during measurement.

Require all original operation/deadline/self-hit/build/reopen gates to pass all
six arms. Do not accept static partition results as a substitute. This remains
a finite research storage component, not full EventFrame or semantic validation.
