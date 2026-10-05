# Bulk construction also passes the768-dimensional finite screen

generation-wide-results.json explicitly records Dimension768, source hashes,
eight per-arm sidecars and retained stores. Its vectors use24 distinct SHA256
blocks per ID, not repeated coordinates. All other gates/caps match the32d run.

All8 arms pass:2048 reads,1024 writes, zero errors or >100ms responses, all seed
self-queries found in top10, all1024 acknowledged records and revisions verified
after reopen. Maximum read latency44.99ms; maximum write latency19.32ms. The30
recorded background builds take67.38-192.34ms; worker drain is included in wall
time. Some runs stop with a nonempty bounded delta rather than forcing a final
compaction. This is valid serving state, not evidence of long-run equilibrium.

| Corpus | Read gap | Max read ms, repeats0 /1 | Max write ms, repeats0 /1 |
| --- | --- | --- | --- |
|200|20ms|8.09 /10.17|16.90 /10.05|
|200|5ms|14.33 /12.47|19.32 /9.94|
|800|20ms|9.73 /11.31|7.63 /10.58|
|800|5ms|13.59 /44.99|9.97 /13.58|

This independently generated vector-width fixture supports the finite component
rescue, not semantic transfer. Width alone does not reproduce real embedding
geometry, adverse clustering, correlated events, million-record search or the
full EventFrame scoring/journal pipeline. The short streams contain128 writes
per arm. Whole-base compaction costs can still outgrow the fixed delta headroom
as the corpus grows, so larger and longer runs are required before promotion.

Next retain the same caps and test larger bases/longer growth. Do not increase
capacity just to move a saturation point outside the measurement window.
All seven whole research goals remain open. Nothing is deployed or pushed.

Verification: `node research/public-task-pilot/check-generation-load.mjs research/public-task-pilot/generation-wide-results.json`.
