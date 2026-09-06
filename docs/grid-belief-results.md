# Fixed-share grid belief: results

Date: 2026-09-06. Parent runtime: d5cda75. Frozen protocol:
[grid-belief-protocol.md](grid-belief-protocol.md). Model parameters were not
tuned between design and confirmation. Raw artifacts contain every trajectory,
the seed base, and all summary intervals. No private data was used.

## What passed and what did not

The narrow range-bias proposition is validated in this synthetic confirmation:
both stationary extremes have positive Bonferroni-adjusted approximate 95%
trajectory-normal Brier gain. Universal improvement is falsified in this fixture.
The stationary .20/.80 controls favor the old exactly matching hypotheses. Beta
is best on all five stationary scenarios. Neither model identifies objective truth.

| Scenario | Old Brier | Grid Brier | Beta Brier | Old minus grid gain | Simultaneous interval | Composed Brier gain |
| --- | --- | --- | --- | --- | --- | --- |
| Stationary .01 | 0.046271 | 0.011679 | 0.009568 | 0.034592 | [0.033975, 0.035209] | 0.014558 |
| Stationary .20 | 0.161864 | 0.167624 | 0.161841 | -0.005760 | [-0.006055, -0.005465] | -0.001750 |
| Stationary .50 | 0.309321 | 0.256933 | 0.251307 | 0.052388 | [0.050415, 0.054361] | 0.000495 |
| Stationary .80 | 0.162235 | 0.168010 | 0.162279 | -0.005774 | [-0.006063, -0.005486] | -0.001767 |
| Stationary .99 | 0.046712 | 0.012493 | 0.010297 | 0.034220 | [0.033627, 0.034812] | 0.014429 |
| Abrupt .99 to .01 | 0.048854 | 0.013522 | 0.250541 | 0.035332 | [0.034678, 0.035986] | 0.014439 |
| Recurring .9/.1 | 0.121374 | 0.111253 | 0.251069 | 0.010121 | [0.009366, 0.010877] | 0.004347 |
| Gradual .1 to .9 | 0.218149 | 0.204567 | 0.250784 | 0.013581 | [0.011863, 0.015300] | -0.001497 |

Each scenario is 64 independent trajectories x 1000 prequential observations.
Intervals use standard error across trajectory means: 1.96 for pointwise 95%,
3.08 for approximate familywise 95% over 24 scenario/metric comparisons. They
are not exact coverage guarantees or confidence sequences. Log loss follows the
raw Brier direction in all eight cases; its full intervals are in JSON.

The fixed composition is .9*.5+.1*belief, with identity calibration and no
residual. On gradual drift its gain interval is [-0.001732, -0.001262], despite
the positive raw-belief gain. This is a concrete warning against promoting a
belief model using a detached loss. No real retrieval or agent answer benefit
was tested. Existing default remains unchanged; grid is opt-in.

## Performance

Apple M4, Go 1.27.0 darwin/arm64, default GOMAXPROCS 10. Same binary, fixed
50-event/32-dimensional hash fixture, recall 50, pack 10. Both two and grid
modes authenticate outcomes. 500 operations x three runs; mixed is 75% recall,
25% durable outcome calls. Includes journal/storage/lock waits; excludes setup
and external signing. No LLM, remote service, or production instance was used.

| Internal workload | Two-hypothesis median ms/op | Grid median ms/op | Two p99 range ms | Grid p99 range ms |
| --- | --- | --- | --- | --- |
| Serial recall | 7.483 | 7.401 | 9.05-9.93 | 9.07-9.88 |
| Serial mixed | 6.581 | 7.038 | 8.98-9.19 | 8.98-9.05 |
| Four-worker recall | 2.489 | 2.654 | 15.02-17.04 | 15.07-22.00 |
| Four-worker mixed | 3.640 | 3.718 | 32.98-35.02 | 33.01-39.00 |

Concurrent ms/op is inverse throughput, not individual latency. Four-worker
mixed cost rose 2.14%; the grid maximum request was 75.00 ms. Run ordering was
not randomized, so small differences are descriptive, not causal estimates.
The older failed 16-worker p99 result remains applicable, not superseded.
Primitive 100,000-update x three runs: median old 440.5 ns, grid 1318 ns;
576 versus 1312 B/op and 6 versus 16 allocations/op. Arithmetic is bounded
O(21), but fingerprinting and allocation remain optimization opportunities.

## Reproduce and checks

```sh
go run ./cmd/grid-belief-experiment -seed 2026090601 > docs/grid-belief-design.json
go run ./cmd/grid-belief-experiment -seed 2026090602 > docs/grid-belief-confirmation.json
go test ./internal/bayes ./internal/service -run '^$' -bench 'Benchmark(GridBelief|WorkingBelief|EvidenceInternalRequests)/(mode=(two|grid))?$' -benchmem -benchtime=500x -count=3 > docs/grid-belief-benchmark.txt
go test ./internal/bayes -run '^$' -bench 'Benchmark(GridBelief|WorkingBelief)$' -benchmem -benchtime=100000x -count=3 > docs/grid-belief-primitive-benchmark.txt
```

The first benchmark regex also matches legacy subtests through its empty match;
all 36 request rows are retained, not selectively filtered. Primitive results
come from the separate command. Full go test, build, vet, and focused race tests
passed. See grid-belief-audit.md for both audit rounds and remaining limits.

## Deployment and possible rescue

Add --grid-belief to --working-belief --evidence-trust-file /path/to/keys.json.
Do not enable alongside hierarchical posterior. Policy changes require fresh
applicable certificates; old working states start at the new uniform prior.
No migration rewrites or data deletion are needed. Grid hypotheses can still
be misspecified, and hazard .02 sacrifices stationary precision for adaptation.

A future rescue may prequentially combine old, grid and stationary experts using
the *composed* proper score and recorded as-of forecasts. It must predict before
updating weights, include selection/admission effects, and pass untouched tests
with baseline/calibration variation. This is a recommendation, not implemented
or validated here; retuning on this confirmation would turn it into design data.
