# Durable boundary timings

The diagnostic completed8 admitted visible-write arms:256 reads and128 writes.
All recorded phase intervals are nonoverlapping and contained within their
operation's arrival-relative duration. Source hashes and successful-record reopen
checks pass. The wrapper delegates each store method exactly once.

## Lower-rate successful operations

Means in milliseconds, pooled across two repetitions.64 reads and32 writes per
packing variant at50reads/s+25writes/s; every operation in these groups succeeds.

| Operation | Total | Dispatch/admission wait | Search | Journal commit | Event put | Uninstrumented remainder |
| --- | --- | --- | --- | --- | --- | --- |
| Original read | 17.248 | 3.712 | 1.220 | 5.842 | 0 | 6.473 |
| Experimental read | 19.392 | 3.850 | 1.915 | 6.079 | 0 | 7.548 |
| Original write | 15.031 | 6.149 | 0 | 0 | 8.832 | 0.051 |
| Experimental write | 17.606 | 8.653 | 0 | 0 | 8.907 | 0.046 |

Journal commit includes serialization, lock acquisition and backend insertion;
event put includes its transaction and index work. Search includes adapter
decoding. These are not measurements of disk-only or CPU-only cost. The
uninstrumented read remainder includes other service/store work and scheduling.

## Higher-rate failure

At200reads/s+100writes/s, original reads have20 errors out of64 and experimental
reads36/64. Both writer groups have20/32 errors. Mean dispatch/admission wait
for failed reads is97.833ms original and96.049ms experimental; for failed writes
it is100.033ms and99.244ms. Most of the error duration is therefore before the
callback's useful work. Successful requests also wait substantially: read means
66.894ms and53.352ms, write means51.078ms and58.404ms.

These conditional means must not replace all-request tail metrics. Some errors
occur after admission inside storage, and cancellation is cooperative. The
diagnostic does not overturn the earlier failed high-rate qualification.

## Research lead, not a confirmed fix

Pinned libravdb1.6.13 source keeps synchronous WAL acknowledgment by default.
Its single-file engine deliberately coalesces commits using a default1ms
adaptive window (maximum5ms, step500us). The explicit target-based5ms option is
only installed when async indexing is enabled; this experiment uses the default
async queue depth0. Do not confuse those two paths or attribute all6-9ms of the
wrapper duration to coalescing.

A useful next ablation is reducing deliberate coalescing delay in an isolated
dependency build overlay while preserving synchronous WAL flush/acknowledgment,
then testing whether lower delay improves admitted workload latency or instead
increases flush contention. This is an untested hypothesis, not permission to
disable synchronization or change installed modules. Another remaining lead is
the6-8ms read remainder; it still needs finer profiling if storage pacing is
insufficient. Production configuration and dependencies remain unchanged.

## Reproduction

DURABLE_PHASE_PROTOCOL.md preceded execution. Raw per-call spans and retained
database paths are in durable-phase-results.json.

```sh
node research/public-task-pilot/check-durable-phases.mjs
```

The full384-operation verifier reports groups by success/error and validates
span containment before calculating the remainder. This is instrumentation
evidence, not a new sustainable-throughput claim. All seven goals remain open.
