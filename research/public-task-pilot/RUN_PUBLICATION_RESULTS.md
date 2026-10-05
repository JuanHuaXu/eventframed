# Retirement-aware consolidation publication

Status: capacity/publication boundary integrated. Sustained-load and complete
serving shutdown remain unvalidated.

`RunLeasePool.AcquireCurrent` and `PublishConsolidation` share the pool mutex.
The latter checks replaced-graph capacity before invoking the writer's atomic
publication. Holding that mutex reserves capacity against competing retirements
and prevents an acquisition from entering between publication and retirement.
After publication, replaced graphs enter the pool's retirement set; zero-reader
graphs are queued and other graphs wait for their final lease.

Capacity rejection leaves the candidate unsettled, allowing explicit abort or
retry. Publication failure follows the candidate's existing cleanup rules. The
caller must use these coordinated methods exclusively and must not bypass them
with raw serving snapshots or direct consolidation publication.

## Tests

`go test -race ./internal/researchindex -run 'TestRunPublication|TestRunLeases' -count=3`
passed in2.192s.

- Replacing two graphs with only one retirement slot is rejected before visible
  state changes; the candidate can still be aborted.
- With sufficient capacity, an old lease continues to return its historical
  answer while a newly acquired current lease returns the replacement answer.
- Releasing old readers eventually closes only the replaced graph.
- The prior bounded/blocked/failed cleanup and shutdown tests remain passing.

The full bulk/candidate-only researchindex suite passed three race-enabled
repetitions in18.670s with the existing research modfile and local overlay.

## Lock and resource contract

Lock order is pool then writer. Persistence callbacks must not call back into
the lease pool or writer. Full manifest planning under publication exclusion can
delay admission; mutex acquisition is not context-selectable. These costs and
tail-latency effects must be measured, not inferred away from passing race tests.

The graph retirement limit still counts handles individually. This integration
does not enlarge it or turn a whole replaced run set into one counted resource.
Published current graphs are not automatically closed by pool shutdown; a final
coordinator must first stop writes/builds/admissions, drain leases and cleanup,
then close the current graph set. Uncertain failed-build cleanup still requires
reconciliation. No production configuration or prior measured artifact changed.

Next run compaction-inclusive, durable load through repeated cycles, recording
admission rejections, deadlines, all construction/planning cost, retirement debt
and final drain. Preserve the original pending-delta and deadline bounds and
explicitly report graph-resource bounds. All seven whole goals remain open.
