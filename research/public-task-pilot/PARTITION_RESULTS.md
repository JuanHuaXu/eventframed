# Fixed partition screen: promising finite result

Protocol: PARTITION_PROTOCOL.md. Artifact: `partition-screen-results.json`.
Runner: `cmd/research-partition-screen`, frozen bulk-build overlay. Four arms,
two repeats of one versus eight partitions; all derived stores retained.

| Layout | Self misses /1024, repeats0/1 | Probe recall@10, repeats0/1 | Replacement ms, repeats0/1 | Median query ms, repeats0/1 |
|---|---|---|---|---|
|One base|1 / 1|0.982813 / 0.982031|1080.79 / 1085.00|1.438 / 1.437|
|Eight partitions|0 / 0|1.000000 / 1.000000|85.18 / 86.84|2.490 / 2.487|

Both eight-partition arms pass all frozen discovery gates. Each has128 random
hash-vector probes evaluated against an exhaustive6400-vector oracle. All4608
queries complete without error. Largest eight-partition query duration4.176ms.
The independent checker recomputes all128 exact oracles, candidate scores,
partition ownership, final merges, hashes and gate outcomes.

Partitions hold757-857 records. Initial per-partition builds83-96ms; the measured
replacement adds32 new records to partition0 and builds an unpublished candidate.
It is NOT a durable update or a concurrent publication test. Only that partition's
payload is rebuilt; partition count and ID assignment remain fixed. Each query
broadcasts to all8 partitions and merges at most80 nominations. Exact partition
merge has200 randomized oracle tests, repeated under the race detector three times.

The tradeoff is about1.73x median query time for about12.5x shorter replacement
build in this fixture. Total ANN effort is not held equal: each partition retains
the backend's400 breadth floor. This may explain part of the recall improvement.
Do not attribute it solely to better graph topology or claim less total query work.

This is a candidate rescue for the observed rebuild bottleneck, not success of
goal6. No durable global revision, cross-partition transaction, lease retirement,
concurrent compaction or full EventFrame integration is implemented by this test.
No claim about semantic embeddings, independent corpora or billion-event scaling.
With fixed8 partitions shard size still grows; increasing partitions is a layout
migration, not a free performance knob. Existing growth failures remain valid.

Next: preserve global publication atomicity while replacing only affected shard
indices; test deletes/updates and old leases across publication, then repeat the
unchanged compaction-inclusive offered load and reopen checks. Do not lower arrival
rates or relax recall to make the new layout pass. All seven goals remain open.
