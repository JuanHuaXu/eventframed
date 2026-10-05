# Committed frontier to actual learner: v1

2026-09-12. Opt-in research-only service tap, disabled by default. No production
deployment, served ranking change, or automatic feedback interpretation.

## Boundary and observed result

The existing ranking callback is deliberately read-only during retries and
clears experimental packet calibration. It was therefore not used to journal
learning. A separate `ResearchFrontierTap` copies bounded compressed features,
baseline probabilities and event bindings after the Bayesian journal commits.
It sees recalled candidates before packing removes them. It does not mean those
candidates were supplied to an agent: the committed and delivered boundaries
remain distinct. Neither full text nor embeddings enter the tap.

An actual Service.Recall test runs 16 requests against three stored fixture
events, with PackK=2. Each successful handoff includes all three candidates,
matching journal ID, snapshot and as-of time. The external test harness journals
all candidate predictions before releasing explicit fixture labels, then sends
those labels to `researchmemory.Background`. It waits for 48 admitted outcomes
and compares all 512 possible feature scores with the synchronous control.
The results match exactly. Complete served packets, including confidence, also
match a tap-disabled service on every request.

This is a service-to-learner lifecycle test using public dummy fixtures and an
in-memory store, not semantic retrieval validation, a useful-answer experiment,
OpenClaw execution, or persistent concurrent performance evidence. Snapshot
compatibility is checked by this harness before learning; automatic dependency
validation and feedback authority are not supplied by the tap.

## Audit and regressions

- Five forced stale journal commits emit zero observations. Three failed attempts
  followed by success emit exactly one observation for that Recall call.
- Tenant isolation, detached output memory, <=200-candidate bound, cancellation,
  closed-tap behavior and queue-full drops are checked.
- Three repetitions of the targeted service research race tests pass, including
  frontier, ranking, shadow and snapshot-interleaving tests. `go vet` passes for
  service and researchmemory.
- The first benchmark exposed extraction before a known-full queue. An early
  capacity/closed check now skips that work; a second locked admission check
  handles competing producers without blocking on queue space.

## Isolated cost

Apple M4, darwin/arm64, 200 candidates with short fixture fields. One repetition
before and one after the early capacity check, not a statistical latency study:

| Mode | Before ns/op | After ns/op | After B/op | After allocs/op |
| --- | ---: | ---: | ---: | ---: |
| Disabled | 5.829 | 5.714 | 0 | 0 |
| Queue drained after each call | 134981 | 135343 | 172929 | 2201 |
| Queue already full | 112710 | 8.291 | 0 | 0 |

Command: `go test ./internal/service -run '^$' -bench BenchmarkResearchFrontierTap -benchmem -count=1`.
These isolate tap preparation/admission, not full Recall or loaded p99. Accepted
work still incurs feature-extraction allocations. Worst-size fields, persistence,
concurrent fitting/serving and realistic backlog still require measurement.

## Remaining boundaries

The consumer owns the tap/worker lifetimes. A bridge must bind verified feedback
to `(journal ID, event ID)` and deduplicate repeated observations across separate
requests; this tap only avoids side effects from failed attempts within a Recall.
The test harness uses distinct as-of requests, not an idempotency proof across
repeated requests. Epoch changes, delayed feedback, stale dependencies and
publication remain explicit future integration work. No learner result affects
the served law or ranking. The previously failed lexical feature representation
is unchanged, so this engineering pass is not an accuracy rescue.
