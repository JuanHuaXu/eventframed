# Compaction-inclusive durable run load: FAIL overall

Two800-record arms PASS. Both3200 and both6400 arms FAIL. Static construction
gains did not eliminate consolidation pauses under the fixed pending-delta cap.

| Initial records | Repeat | Write errors | Read errors | Late | Self misses | Ack/reopened | Consolidations |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 800 | 0 | 0 | 0 | 0 | 0 | 512/512 | 8 |
| 3200 | 0 | 98 | 0 | 0 | 0 | 414/414 | 6 |
| 6400 | 0 | 288 | 0 | 0 | 1 | 224/224 | 3 |
| 800 | 1 | 0 | 0 | 0 | 0 | 512/512 | 8 |
| 3200 | 1 | 100 | 0 | 0 | 0 | 412/412 | 6 |
| 6400 | 1 | 287 | 0 | 0 | 1 | 225/225 | 3 |

All773 write errors are `research index delta full`. All2299 acknowledged writes
reopen, with no build/audit errors or failed-present writes. Two successful-query
self misses at6400 remain an ANN nomination failure signal; this experiment does
not establish their precise graph-level cause. No successful-response deadline
exceedances occurred, but rejected writes still fail the service gate.

Flushes take21-44ms. Consolidation takes102-187ms at800,484-547ms at3200 and
1310-1389ms at6400. One builder means no other delta drain during those long
consolidations. Headroom from32 to64 pending records at100 writes/s is about320ms,
which the larger consolidations exceed. Making consolidation less frequent alone
does not remove that finite-capacity stall.

Final lease cleanup plus current-graph close takes92.5-243.1ms, recorded outside
request timers. All requests/builders are joined before shutdown and authority
reopen. The shutdown helper is harness-only and requires external reader drain.

## Artifacts and verification

`run-load-results.json`, six per-arm sidecars and private `.stores` directories.
`RUN_LOAD_PROTOCOL.md`, `build-run-load.mjs`, `cmd/research-run-load/main.go` and
`check-run-load.mjs`. The generic verifier accounts for hashes, sidecars and9216
operations; the run-specific verifier additionally checks layout transitions and
at least two successful consolidations per arm. Verifier process success means
consistent artifacts, NOT passed performance gates; the table records failures.

## Interpretation and next lead

This layout uses one whole initial graph instead of eight ID partitions. It
does not dominate the earlier partitioned design, which has smaller individual
rebuild pauses. With two retired-graph slots, consolidating a pair of runs fits
the frozen resource count, but still rebuilds the whole corpus. No capacity or
deadline gate was relaxed and no adoption follows this failed experiment.

Next examine bounded sub-corpus consolidation, or permitting a bounded small-run
flush while consolidation is in flight. The latter must retain post-capture runs
at publication, account for the extra unpublished graph/CPU work explicitly, and
avoid presenting increased buffering as a speedup. Either direction must rerun
the same arrival rates, pending cap, durability and recall checks. Full EventFrame
semantic validation is still separate. All seven research goals remain open.
