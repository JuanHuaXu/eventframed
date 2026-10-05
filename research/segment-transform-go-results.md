# Go tail-transform component pilot

Research-only test files: `internal/observationlearners/segment_transform_research_test.go`
and `segment_transform_check_test.go`. The frozen fitter and serving code are
unchanged. This implements the algebra in `segment-transform-lead.md`, not a
new learner or a rescue of its failed quality criteria.

## Verification

Compared all likelihood entries exactly and 783,360 tail forecasts within
1e-12, for 0/1/16/64 samples, three fixtures (existing ridge fixture, repeated
all-positive input, alternating labels), and family masses
0, 1e-12, .5, .95, 1-1e-12, 1. Maximum observed absolute forecast difference:
9.1038288019262836e-15. No range violations. This is numerical, not bitwise,
forecast equivalence. Full posterior fitting/integration is not yet tested.

## Component measurement

Apple M4, darwin/arm64. Sequential benchmarks, 300ms target per case; no loaded
serving or p99 measurement. Final scratch-reuse run:

| Labels | Original ms/op | Transform ms/op | Original B/op | Transform B/op |
|---|---:|---:|---:|---:|
| 16 | 7.654252 | 5.949650 | 303104 | 466946 |
| 64 | 63.513242 | 56.477298 | 303104 | 466944 |

Approximately 22% and 11% lower component time, respectively. Allocation count
is 1 versus 2. The transform still adds about 160 KiB scratch allocation per
fit. All cell-mask decoding is included; existing mask/prior lookup tables are
warm/shared in these paired measurements, so this is not a cold-start test.

An initial version allocated scratch per tail: 17/65 allocations and about
2.92/10.79 MB per fit at 16/64 labels. Three initial benchmark repetitions gave
original ranges 7.58-7.70/62.56-63.58ms and transform ranges
5.89-5.92/56.20-56.56ms. Fit-local scratch reuse removed that avoidable allocation
growth; every cell is overwritten before each transform, avoiding stale state.

Reproduce final checks:

```sh
go test ./internal/observationlearners -run '^TestSegmentTransformParity$' -v -bench '^BenchmarkSegmentTransformPair$' -benchtime=300ms -count=1
```

This leaves direction 6 open. The interval likelihood recurrence is unchanged
and remains quadratic in label count; accelerating the tail alone cannot remove
its cost. Next steps: full-posterior equivalence and ownership tests, then a
separate loaded background-fit experiment. No production adoption or latency
guarantee follows from this component pilot.
