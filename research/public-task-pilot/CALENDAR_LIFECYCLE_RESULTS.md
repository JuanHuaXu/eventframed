# Calendar lifecycle and component cost

Actual service lifecycle tests passed three repeated runs under Go's race
detector. Each repetition uses fresh in-memory stores and hash embeddings;
no production resources or external model calls.

-64 concurrent recalls, two tenants, opposite exclusion queries that collapse
 to the same focused query: all correct tenant-specific winners, pack1.
 Across three repetitions this is192 requests, not192 independent task examples.
-A barrier after rank computation followed by forced policy-version motion:
 journal validation rejects the stale attempt; exactly two ranker calls; the
 returned packet carries the new snapshot and correct winner.
-Cancellation while blocked after ranking returns context.Canceled and no packet.
-Forced version motion on every attempt exhausts exactly five attempts, returning
 ErrStaleSnapshot and no packet rather than publishing a stale result.

The policy mutation is a test instrument used to exercise the real store's
journal snapshot check. It is not an experiment in posterior learning, concurrent
LibraVDB writes or production configuration mutation. The adapter remains
stateless and read-only; only test wrappers mutate or block.

Reproduce:

```sh
go test -race ./internal/researchcalendar -run TestService -count=3 -v
```

## Component Benchmark

Go1.27.1, darwin/arm64, Apple M4, GOMAXPROCS4, five benchmark repetitions,
200ms per benchmark calibration target, no concurrent experiment processes.
Values below are medians of five benchmark averages, NOT p95/p99 latencies.

| Frontier | Passthrough | Calendar adapter | Calendar heap/op | Allocs/op |
| --- | --- | --- | --- | --- |
|50 |0.414 us |67.118 us |14583 bytes |234 |
|200 |1.496 us |217.063 us |52990 bytes |834 |

The component includes original-query binding validation/canonicalization,
date parsing, duplicate-ID/score validation, stable partition and ordinal-score
assignment. It has no I/O or learned updates. Setup is excluded from the timer.
The frontier repeats two real Rosetta facts under distinct fixture IDs solely
to exercise50/200 records; it is not independent evidence or an accuracy test.
The control copies candidate order and does not perform equivalent reasoning.

Runtime grows with input text and frontier size; capped candidate count does not
make scanning arbitrary text free. This implementation caps original query at
4096 bytes, each candidate text at8192 bytes and frontier at200. Its normal
parsing/partition work is linear in total input length plus expected hash-map
work, with allocation proportional to frontier size. Maximum-size text, allocation
pressure under sustained traffic, and end-to-end tail latency remain unmeasured.

```sh
go test ./internal/researchcalendar -run '^$' -bench BenchmarkCalendarFrontier -benchmem -benchtime=200ms -count=5 -cpu=4
```

Raw output: [calendar-frontier-benchmark.txt](calendar-frontier-benchmark.txt).
No claim that217us equals total serving overhead under real load or that these
results complete direction6. Next unresolved integration questions: later learned
deltas changing ordinal ordering, durable reasons and calibration semantics,
and realistic retrieval/persistence load. All seven whole directions remain open.
