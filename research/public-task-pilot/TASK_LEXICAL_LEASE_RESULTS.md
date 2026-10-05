# Request-lease scheduling: finite rescue

## Result

The corrected harness completed1024 reads and512 writes across32 arms. With
leases, all512 reads and256 writes succeeded; there were zero stale journal
rejections and no future evidence in successful returned frontiers. Without
leases,14 reads failed, all in experimental visible-write conditions. The earlier
run's17 failures and miscounted215 harness flags remain in their original artifact.

| Experimental visible-write arm | Unleased p95 ms, runs1/2 | Leased p95 ms | Unleased failures | Leased failures |
| --- | --- | --- | --- | --- |
| 50 candidates | 47.246 / 46.548 | 10.501 / 10.674 | 3 / 3 | 0 / 0 |
| 200 candidates | 160.962 / 159.191 | 30.080 / 30.638 | 4 / 4 | 0 / 0 |

The finite no-error/no-stale/no-future-leak gate passes. Across all leased arms,
the largest reader duration was47.193ms and largest writer duration24.841ms,
both including admission wait. These observations satisfy the recorded100ms
check, not a general or production latency guarantee.

## Tradeoff, not free concurrency

For the experimental200-item visible-write case, total completion time for the
same32 reads+16 writes fell from374.617/402.150ms to282.075/284.261ms. Maximum
writer duration rose from0.628/1.204ms to23.996/24.194ms. The reader improvement
does not hide writer wait: it is measured and reported explicitly.

Compatible future writes are also blocked by this coarse prototype. This adds
unnecessary writer latency and sometimes reader queuing to a condition that
already needed no retries. Leases should not be described as universally faster.
The prototype trades optimistic re-execution for coordinated admission.

## Mechanism and boundary

The research runner uses the existing x/sync weighted semaphore, with four
permits: each Recall holds one through its full call; each CaptureTurn writer
holds all four. Acquire is context-aware. Durations start before acquisition,
and permits are released on ordinary success/error return. The service's
snapshot validator, retry count and calendar/lexical algorithms are unchanged.

This is **runner-level scheduling**, not an implementation in eventframed's
public daemon API. All mutations in this fixture participate. Uncoordinated
deletion, graph/policy updates, other processes or tenants are not covered.
Queued-admission cancellation is not specifically exercised by this successful
load run; library support alone is not an end-to-end cancellation test.

The finite workload is closed-loop: one writer sleeps5ms after each completed
write. Leases therefore change actual write arrival timing, although total work
is unchanged. No fairness or stability guarantee under open-loop overload follows.
The memory store, hash embedder and repeated public facts also remain narrower
than durable backend and agent workloads.

## Verification and continuation

TASK_LEXICAL_LEASE_PROTOCOL.md preceded dispatch. Raw per-request/writer durations,
waits, counters and hashes are retained in task-lexical-lease-results.json.

```sh
node research/public-task-pilot/check-task-lexical-lease.mjs
```

Runner cmd/research-task-lexical-lease requires researchpriority and the existing
task-lexical-overlay-v1 Go overlay, then a NEW artifact path. Source inspection
and observations distinguish actual Recall failures from nomination bounds.

Next: make admission an explicit research component with cancellation and
participant-boundary tests, then test durable writes and open-loop arrivals.
Consider snapshot-aware admission to avoid blocking provably compatible future
ingestion, but do not skip synchronization based on a racy time check. All seven
whole directions remain open; no production or whitepaper changes.
