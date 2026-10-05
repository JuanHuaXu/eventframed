# Partition view/index publication and leases

`internal/researchindex/partition_serving.go` now wraps the partitioned durable
coordinator with matching immutable HNSW handles. This remains research-only.

## Implemented Boundary

- Acquire captures one coherent global view and copies its matching handle set
  under serving exclusion. External compaction is forbidden and mismatches fail.
- Queries search all fixed partitions and merge their local top-k after each
  partition applies its own delta shadowing/tombstones. Errors do not yield a
  partial successful result.
- Compact prebuilds one replacement graph outside admission, then publishes its
  rebased view and matching handle while acquisitions are excluded. Other handles
  stay unchanged; intervening durable writes survive in the selected delta.
- Leases pin all required handles. Release drops vectors/handle references while
  preserving the scalar historical revision. The lease limit counts whole queries.
- The retirement limit counts individual old graphs globally, not per partition.
  Closing handles remain charged until success. Close refuses active readers,
  writes, builds or unresolved retired resources.

## Tests And Bug Hunt

Three full researchindex race-enabled repetitions pass with the bulk-build
overlay. New tests cover historical readers, exact handle ownership, revision
preservation, lease/retirement limits, in-build updates and deletes, failed builds,
stalled close accounting, and100 searches overlapping20 writes/compactions.

Review identified a real cleanup gap in the new wrapper: if cancellation prevented
publication and closing the unpublished graph returned an error, that error was
discarded and the graph was no longer charged. The added regression failed with
`cleanup error hidden: context canceled`. Compact now joins the cleanup error and
charges the unpublished graph until successful reclamation. The regression and
full suite pass. The test injects a persistent close-error state after closing
its real fixture DB, so it does not intentionally leak storage resources.

The frozen single-base research control has the analogous discarded cleanup
pattern; it is preserved for historical reproducibility, not endorsed for
production use. The new wrapper does not inherit that behavior.

## Remaining Load Question

Next run the existing6400/3200/800 growth schedule: delta64 TOTAL, trigger32 TOTAL,
one background builder, two retired handles, eight leases, no request retries.
Choose the partition with the largest current delta, deterministic index tie-break.
Do not multiply the budget by eight or change the arrival rates.

A partition rebuild may remove fewer than32 entries. Therefore the85-87ms static
replacement result does not by itself prove background service capacity. Measure
actual removed entries, compaction durations, admission errors, due-relative tail
latency and durable reopen results. Full EventFrame contracts and all seven whole
research goals remain open; there is no production integration or new runtime claim.
