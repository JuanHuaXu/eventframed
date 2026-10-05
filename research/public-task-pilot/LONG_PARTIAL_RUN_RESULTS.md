# Long partial-tail growth: FAIL in both repetitions

The eight-times-longer workload falsifies sustainable throughput for the current
single-growing-tail merge policy. The earlier512-write pass remains valid only
for its finite interval; it is not a production-capacity result.

| Repeat | Read errors | Write errors | Late reads | Self misses | Ack/reopened | Partial merges |
| --- | --- | --- | --- | --- | --- | --- |
| 0 | 52 | 149 | 0 | 0 | 3947/3947 | 96 |
| 1 | 79 | 136 | 25 | 0 | 3960/3960 | 95 |

All285 write errors are pending-delta capacity rejections. First rejected write
indices are2996 and3068 (zero-based). All7907 acknowledged writes survive reopen;
no build/audit errors or failed-present writes. The mode-aware verifier checks
hashes, sidecars,24576 operations, flush transitions, partial merge transitions
and zero delta removal by partial merges. Both performance gates fail.

## Growth and admission evidence

First partial merges take30.54/29.27ms; last merges557.72/791.12ms; maximum
823.11/807.11ms. The first >320ms partial merge begins near durable revision1938
in repeat0 and1871 in repeat1. Unlike flush, a partial merge can begin with very
little pending data, so its actual headroom may exceed320ms; the invariant is
available capacity divided by arrival rate, not a universal320ms threshold.
Eventually merge duration exceeds even near-empty-buffer headroom at100 writes/s.

Repeat0 read errors are52 reader-admission refusals. Repeat1 has57 admission
refusals and22 context-deadline errors. Its25 late reads reach185.02ms; some
deadline errors can overlap that late count and must not be added as distinct
failures. Repeat1's first read error occurs at index519, well before the first
capacity rejection, so not all admission failures can be attributed to a grown
tail from these aggregate logs. Detailed scheduling/lock tracing would be needed.
No writes exceeded100ms. Final cleanup/drain takes467.48/463.32ms.

## Artifacts

`long-partial-run-results.json`, two per-arm sidecars and `.stores` directories;
`LONG_PARTIAL_RUN_PROTOCOL.md`; `build-long-partial-run.mjs`;
`cmd/research-long-partial-run/main.go`; `check-long-partial-run.mjs`.
The original short experiment and its verifier remain unchanged.

## Next direction

Repeatedly merging the accumulated tail with the latest small run creates growing
rewrite work. A size-tiered or sub-corpus policy should avoid that repeated large
merge, but still needs a bound on the largest nonpreemptible build and retained
resources. Merely delaying consolidation or increasing pending capacity is not
a demonstrated rescue. Before another loaded screen, model the merge schedule's
worst build size/headroom and test tombstone/version preservation for the selected
subsets. Investigate the separate admission/deadline cluster rather than claiming
all failures share one proven cause. Update/delete load and full EventFrame agent
semantics remain untested here. All seven whole research goals remain open.
