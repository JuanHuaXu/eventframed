# Trace attribution and residual-dispatch candidate

## Evidence

The execution-trace/heap diagnostic completed256 checked recalls,128 per arm.
All nominations and journals match; raw trace and profile source hashes verify.
The heap comparison subtracts the pre-loop baseline. Experimental sampled
allocation is1280.53MB, distributed across string tokenization, epistemic terms,
candidate copies, FrameText construction, JSON decoding and retrieval metadata.
No single allocation site justifies replacing the data model wholesale.

The experimental scheduler-delay profile totals86.98ms across goroutines;
loadResidualCandidates accounts for62.02ms cumulative (71.30%). The ordinary
profile totals91.15ms, with66.21ms (72.64%) in that loader. These are scheduling
delays, not CPU percentages or additive per-request wall costs. They refine the
earlier CPU lead: the residual loader's channel dispatch is more directly
implicated than shard-search spawning, but do not uniquely explain every CPU
sample in runtime.pthread_cond_signal.

The synchronization profile attributes773.20ms cumulative blocking to journal
insertion over128 experimental calls, consistent with the earlier approximately
6ms journal boundary timing. Background and trace goroutine waiting contributes
to the profile total and must not be called request latency.

## Fixed-stride candidate: useful counterexample, not adopted

An unwired research alternative assigns candidate indices to the same maximum
eight workers using fixed strides, instead of dispatching each through a channel.
It preserves indexed results, missing-frontier validation, disabled behavior,
error propagation, context cancellation and joining before return in the focused
race-enabled tests. Candidate counts tested:0,1,7,8,9,50,200.

An isolated cheap-lookup200-candidate benchmark on Apple M4/CPU4, three200ms runs:

| Dispatch | ns/op runs | Allocations/op |
| --- | --- | --- |
| Existing channel | 97377,96981,95246 | 1017 |
| Fixed stride | 37236,36896,37007 | 1023 |

The median improvement is2.62x for this helper, not full service latency.
However, a deterministic virtual-time counterexample assigns1ms latency to
every eighth key: existing dispatch completes in4ms, fixed stride in25ms,
with identical results. Slow keys concentrate on one stride. The passing
counterexample test documents a performance failure, not optimization success.

Do not adopt fixed strides as a general replacement. Next test an atomic
next-index allocator: retain dynamic work balancing while removing per-item
channel handoffs. It must pass both uniform-cost and skewed-cost comparisons,
cancellation/error tests and actual service output/load tests before promotion.
No production path calls either a new dispatcher or a weaker residual model.
All seven full directions remain open.

## Reproduction

Raw data: durable-trace-results.json and its trace/heap suffixes. The exporter
verifies inputs and exclusively creates scheduler/synchronization profiles;
do not rerun it over existing derivatives.

```sh
go tool pprof -top -cum research/public-task-pilot/durable-trace-results.json.true.trace.sched.pprof
go tool pprof -top -alloc_space -base research/public-task-pilot/durable-trace-results.json.true.trace.before.heap research/public-task-pilot/durable-trace-results.json.true.trace.after.heap
go test -race ./internal/service -run '^TestResearchResidualStride' -count=1 -v -timeout=120s
go test ./internal/service -run '^$' -bench '^BenchmarkResearchResidualStride$' -benchmem -benchtime=200ms -count=3 -cpu=4
```
