# Partition-base / partial-tail load: six arms PASS

All six frozen arms pass: initial800/3200/6400 records, two fresh repetitions,
6144 reads and3072 writes total. No request errors, no >100ms responses, no
successful self-query misses, no build/audit errors and no failed-present writes.
All3072 acknowledgements survive authoritative close/reopen with matching revision.

| Initial records | Repeat | Partial merges | Drained delta entries | Ack/reopened |
| --- | --- | --- | --- | --- |
| 800 | 0 | 14 | 480 | 512/512 |
| 3200 | 0 | 15 | 512 | 512/512 |
| 6400 | 0 | 14 | 512 | 512/512 |
| 800 | 1 | 14 | 480 | 512/512 |
| 3200 | 1 | 14 | 480 | 512/512 |
| 6400 | 1 | 14 | 512 | 512/512 |

Undrained final entries remain durably committed; the harness does not silently
add a final flush. Partial merges drain zero pending entries, verified from
before/after revisions and delta counts. Flushes take16.89-47.15ms; partial merges
26.43-75.75ms. Final lease/current-graph drain takes109.88-312.34ms outside request
timers. These are finite loaded observations, not tail probability guarantees.

## What changed

Eight fixed base partitions are queried with up to two newer tail runs. Merge
only the two newest tail runs; leave base graphs unchanged. A partial merge
preserves tombstones because older unmerged history still exists, and retains
the entire pending delta (including pre-capture entries). Its publication uses
the lease pool's two-graph retirement reservation. A real-graph concurrent test
checks tombstone retention, pre/post-capture delta preservation and unchanged
older graph identity. Three targeted race repetitions passed in1.874s; three
full bulk/candidate-only suite repetitions passed in19.069s before measurement.

The failed whole-history design incurred773 capacity rejections. This layout
avoids its large-corpus rebuild pauses in the measured interval. It also uses
eight initial partitions rather than a whole-base graph, so the comparison does
not isolate partial merging as the sole cause of improved recall or throughput.

CPU4,64 pending records,8 reader leases,2 retired graph slots and100ms deadlines
remain fixed. Current graph count increases from at most2 in run-load to at most10,
plus2 retired and1 candidate. This is an explicit representation/resource change,
not evidence of equal peak RAM. No private data, production or normal module changes.

## Audit artifacts

`partial-run-load-results.json`, six per-arm sidecars and `.stores` directories;
`PARTIAL_RUN_LOAD_PROTOCOL.md`; `build-partial-run-load.mjs`;
`cmd/research-partial-run-load/main.go`; `check-partial-run-load.mjs`.

The original generic verifier rejected this output because it assumes every
build starts with at least32 pending entries. That assumption is inapplicable to
the predeclared partial-merge trigger. It remains unchanged. The new verifier
requires >=32 for flushes, requires10->9 graph transitions for partial merges,
and requires zero delta removal for each successful partial merge, while retaining
hash, sidecar, operation, durability, deadline, error and recall checks. All pass.

## Remaining falsifier

Each arm lasts only512 writes. The merged tail grows as data accumulates, even
though graph count is bounded. Longer runs must reveal whether its build time
eventually exceeds the320ms pending-buffer headroom. Base updates/deletes and
ownership-plan construction can also raise costs beyond append-only behavior.
Next test longer growth and update/delete workloads with explicit memory/candidate
budgets before claiming sustainable throughput. Full EventFrame/agent semantics
remain a separate goal. No whole research goal is marked complete.
