# Cancellation boundary before shadow integration

Inspection confirmed that `ResearchShadowPolicy.Process` must honor context
cancellation, while the frozen segment fitter and first transform candidate
accept no context. Wiring either directly would violate that processor contract.
The ordinary `researchmemory.Background` instead fits its separate adapter;
it is not currently a segment-fitter integration. Do not conflate those paths.

Added a research/test-only context-aware copy. Checks occur before fitting,
between label intervals, between boundary-recurrence steps, and before returning
a completed model. Cancellation returns nil model plus context error; no
partial posterior is published or persisted. Fit-local arrays are not retained.
No production or frozen source was changed.

`TestTransformCancellation` passed with race instrumentation: the maximum
fixture traverses2420 checkpoints. Eight deterministic cancellation cutoffs,
including near-return cutoffs, all reject partial output. A real1ms deadline
also rejects output. An uncancelled fit exactly matches the non-context
transform; refitting after cancellation produces the same model. Input history
is unchanged. The fake checkpoint context is single-threaded and tests control
flow, not a concurrent implementation of the context interface.

Cancellation is cooperative, not preemptive: a checkpoint cannot interrupt a
running mask/tail loop or a runtime scheduling pause. No hard wall-clock bound
is established. The service must still recheck deadline/snapshot before using
a result, as its existing shadow worker does.

Three sequential300ms benchmark targets on Apple M4:

| Variant | ms/op across three runs | B/op | allocations |
|---|---|---:|---:|
| No context |57.699,58.534,58.125|477184|7|
| Active deadline |57.779,57.633,57.609|477184|7|

No observable slowdown in this small component sample; do not claim a speedup
from adding cancellation. Timer construction is outside the timed loop, so
this does not measure per-job scheduler/context creation cost.

```sh
go test ./internal/observationlearners -run '^TestTransformCancellation$' -race -count=1 -v
go test ./internal/observationlearners -run '^$' -bench '^BenchmarkTransformContextPair$' -benchtime=300ms -count=3
```

Actual shadow-scheduler/foreground interference remains the next experiment.
These checks establish a prerequisite, not direction6 success.
