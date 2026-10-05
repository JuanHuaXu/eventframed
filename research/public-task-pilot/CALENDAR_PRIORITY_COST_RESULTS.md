# Priority lifecycle and cost rescue

The exact service overlay passed three race-enabled repetitions of64 concurrent
opposite-exclusion requests (192 total), checking winner, session, journal
explanation and snapshot binding. Injected journal storage failure and a store
returning context.Canceled each return an error and empty packet, without a
second commit attempt. The latter tests error propagation, not a live cancellation
race or proof that a real backend honors cancellation.

## Cost diagnosis

The priority hook was measured against ordinary packing, including plan creation,
FrameText projection and explanation JSON, excluding embedding, journal writes,
network, fitting and the rest of Recall. Both use diversity packing, pack10,
frontier50/200, budget10000. Two repeated public Rosetta facts are cardinality
fixtures, NOT independent observations. Apple M4, darwin/arm64, cpu4, three
200ms Go benchmark samples; medians of average ns/op, not p95/p99.

| Candidates | Adaptive | Ordinary packing ms | Original priority hook ms |
| ---: | --- | ---: | ---: |
| 50 | off | 0.630 | 0.720 |
| 50 | on | 1.843 | 3.783 |
| 200 | off | 2.806 | 3.069 |
| 200 | on | 9.542 | 19.153 |

Inspection confirms the extra work: PackPriority runs full packing to obtain its
Expanded flag, then runs full packing again with selection-local priority.
The first pass repeats diversity and evidence work needlessly. The actual
shouldExpand decision uses **forecast-law margin/entropy and event priority**,
not numeric rank scores. Earlier notes calling this original-score expansion
were imprecise; original candidate order and forecast fields are what matter.

## Separate fast variant

An additive research helper exposes the existing shouldExpand decision and the
same cap arithmetic without selection. PackPriorityFast uses it; original
PackPriority and previously hashed service sources stay untouched. This is a
component candidate, NOT yet substituted into the full-service overlay.

A deterministic300-case randomized equivalence test passed under the race
detector: complete packing results and byte-identical explanation strings match
the old hook across varying laws, scores, priorities, budgets, pack limits,
adaptive caps and diversity settings. No change to calendar predicates or
acceptance criteria was made to obtain speed.

Paired follow-up benchmark:

| Candidates | Adaptive | Original hook ms | Fast hook ms |
| ---: | --- | ---: | ---: |
| 50 | off | 0.724 | 0.721 |
| 50 | on | 3.782 | 1.916 |
| 200 | off | 3.101 | 3.059 |
| 200 | on | 19.201 | 9.493 |

Adaptive cost falls approximately49-51% in this fixture. Without adaptive
expansion, no substantial improvement is established. The largest fast fixture
still allocates roughly1.93MB/op, much of it inherited packing work. These are
not loaded service timings and do not establish a sub100ms serving guarantee.

## Evidence and next step

Raw output: calendar-priority-hook-benchmark.txt and
calendar-priority-fast-benchmark.txt. New test files are
internal/researchcalendar/priority_failure_overlay_test.go and
internal/service/calendar_priority_fast_overlay_test.go. Benchmark drivers live
in the service package with matching calendar_priority names.

```sh
go test -race -tags researchpriority -overlay research/public-task-pilot/priority-overlay-v1/overlay.json ./internal/researchcalendar -run 'TestPriorityOverlay(JournalFailure|ConcurrentJournalBinding)' -count=3 -v
go test -race -tags researchpriority -overlay research/public-task-pilot/priority-overlay-v1/overlay.json ./internal/service -run TestPriorityFastHookEquivalence -v
go test -tags researchpriority -overlay research/public-task-pilot/priority-overlay-v1/overlay.json ./internal/service -run '^$' -bench BenchmarkPriorityOverlayHook -benchmem -benchtime=200ms -count=3 -cpu=4
go test -tags researchpriority -overlay research/public-task-pilot/priority-overlay-v1/overlay.json ./internal/service -run '^$' -bench BenchmarkPriorityFastHook -benchmem -benchtime=200ms -count=3 -cpu=4
```

Next integrate this equivalent variant into a new overlay, repeat actual service
checks, and measure realistic load. Absent/ambiguous-answer handling and fresh
agent-usefulness validation remain required. No whole direction is complete.
