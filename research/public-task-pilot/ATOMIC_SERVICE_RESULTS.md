# Atomic residual allocation: helper success, durable rescue failure

## Helper and behavior

The research allocator uses one atomic next-index counter with at most eight
workers. It retains dynamic work allocation, unique result indices and joining
all workers before return. No queue or worker survives the request. The original
channel and rejected fixed-stride implementations remain unchanged as references.

Focused race-enabled tests pass five repetitions: candidate counts0/1/7/8/9/50/200,
output parity, disabled/missing-frontier behavior, lookup errors and cancellation
with eight blocked workers. In the skewed virtual-time fixture, original and
atomic dispatch both complete4ms; the earlier fixed-stride variant took25ms.
This is not proof of optimal balancing under arbitrary lookup latency.

Cheap-lookup200-candidate microbenchmark, Apple M4/CPU4, three200ms runs:

| Variant | ns/op runs | allocations/op |
| --- | --- | --- |
| Original channel | 97729,98601,96152 | 1017 |
| Atomic index | 36344,35830,35867 | 1016 |

The median helper speedup is2.72x. Allocation volume is about143KB/op in both.
This isolated result does not imply a similar service speedup.

## Service replay

All108 public outputs equal saved incremental results after removing timing
only, including scores, laws, rank traces, support ranks and explanations.
Focused overlay tests across packing, researchcalendar and service pass under
the race detector. The datasets remain consumed design tests; their known
wrong answers and absent-answer behavior are unchanged.

## Durable fixed-arrival screen

All32 arms completed. Every lower-rate arm passes; every higher-rate arm fails.
The full durable rescue is therefore rejected despite the helper improvement.

| High-rate admission | Prior read errors /256 | Atomic read errors /256 | Prior write errors /128 | Atomic write errors /128 |
| --- | --- | --- | --- | --- |
| off | 137 | 134 | 32 | 26 |
| on | 110 | 106 | 79 | 84 |

The comparison is cross-run, not paired simultaneous measurement. Small count
differences do not establish improvement or regression statistically. Admitted
conditions have no stale journal rejections, but deadlines still expire.
Across1024 reads and512 writes there are350 errors and98 callbacks never entered.
All784 returned journals and402 acknowledged writes survive orderly reopen;
warmups pass and successful frontiers contain no observed future evidence.

The scheduling overhead reduction is insufficient to overcome the durable
workload's remaining serialized write cost and queueing. Do not promote this
as a throughput rescue or claim that durable persistence became faster.

## Next lead and audit

Test bounded transactional write batching as a separate research design: can
several already-pending events share durable work while preserving individual
idempotency, final snapshot publication, cancellation outcomes and synchronous
acknowledgements? First inspect existing batch APIs and scope the linearization
contract. Dropping writes, disabling synchronization or moving acknowledgements
ahead of commit would change the requirement rather than solve it.

ATOMIC_SERVICE_PROTOCOL.md preceded service runs. atomic-overlay-v1 alters only
the residual-loader call on top of incremental diversity. Normal modules and
production remain unchanged. Raw load: atomic-durable-results.json. Public
outputs: each set's atomic-service-results.json.

```sh
node research/public-task-pilot/check-atomic-service.mjs
node research/public-task-pilot/check-atomic-durable.mjs
go test -race ./internal/service -run '^TestResearchResidualAtomic' -count=5 -timeout=120s
```

All seven whole research goals remain open.
