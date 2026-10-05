# Durable steady-state CPU diagnostic

The default-module, default-WAL experiment completed128 ordinary and128
experimental recalls on separate fresh local libravdb databases. All256 returns
have200 nominations and matching journal explanations. Source/profile hashes
were checked. The incremental diversity overlay was active; the rejected
zero-window dependency variant was not used.

## Samples

| Sampled CPU attribution | Ordinary | Experimental |
| --- | --- | --- |
| Total CPU samples | 1.02s | 1.22s |
| runtime.pthread_cond_signal, flat | 520ms (50.98%) | 590ms (48.36%) |
| runtime.madvise, flat | 80ms (7.84%) | 190ms (15.57%) |
| Service.Recall, cumulative | 170ms (16.67%) | 190ms (15.57%) |

These short profiles show runtime wakeup/memory-management work rather than the
earlier repeated diversity comparisons. Cumulative stacks overlap; percentages
must not be added. CPU profiling omits off-CPU I/O and blocking, and runtime
systemstack frames do not by themselves identify the initiating application
call site. This is not evidence that pthread operations are a code bug.

The128-call arm wall times are1.683s ordinary and1.907s experimental. Cumulative
allocation is1,177,629,544bytes and1,332,578,432bytes respectively, including the
profile/readback interval. This is allocated volume, not resident or retained
memory. It motivates allocation-site measurement, not disabling garbage collection.

## Next discriminating evidence

Pinned collection search starts goroutines for shard searches and a completion
waiter. That is one credible source of wakeups, but these profiles do not prove
that it dominates. Storage completion and runtime allocation also wake work.
An execution trace should attribute goroutine creation and scheduling latency;
an allocation profile should identify repeated materialization. Do not replace
parallel search or tune GOMAXPROCS from these CPU samples alone.

The durable high-rate failure remains unresolved. These read-only profiles have
no event-write contention, admission queue or long-duration overload, and cannot
qualify production throughput. All seven whole goals remain open.

## Artifacts

DURABLE_CPU_PROTOCOL.md preceded dispatch. JSON: durable-cpu-results.json.
CPU profiles are named with .false.cpu.pprof and .true.cpu.pprof suffixes.

```sh
go tool pprof -top -cum research/public-task-pilot/durable-cpu-results.json.true.cpu.pprof
go tool pprof -top -cum research/public-task-pilot/durable-cpu-results.json.false.cpu.pprof
```

Fresh public-fixture databases are retained under durable-cpu-results.json.stores.
No production changes or whitepaper promotion were made.
