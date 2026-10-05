# Short-window rate model v78 results

Overall: FAIL in both phases. Shortening the rate model addresses much of the
lag diagnosed in v77 and materially improves sparse detection, but it increases
homogeneous misses and loses the long model's weak-signal detections. No serving
change or completed research direction follows.

[Frozen protocol](mmm-short-rate-v78-protocol.md),
[raw artifact](mmm-short-rate-v78.jsonl),
[evaluator](../../research/short-rate-v78-summary.mjs),
[implementation](../../internal/observationgate/window_bet.go).

The experiment contains 10,240 fresh paired streams with 22 source/evaluator
hashes. SHA256:
`11c198098ee7f671bc7ce281e57677ed35322abe0ef0d98cbd155c35d7316d40`.
Eight rate observations per channel are frozen before generating either phase.
Allocation, augmentation, thresholds, starts and query budgets remain unchanged.
No confirmation tuning or retrospective success-threshold changes were made.

## Confirmation

512 trajectories per scenario. Restricted delay charges undetected changes at
the remaining horizon. All comparisons below use the same fresh experiment,
not control values copied from v76.

| Scenario | Uniform delay | Fixed augmented | Long rate | Short rate | Short gain vs uniform | Uniform / long / short misses |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Homogeneous128 | 153.66 | 155.56 | 181.21 | 158.21 | -2.96% | 0 / 3 / 7 |
| Sparse128 | 137.62 | 127.95 | 117.60 | 89.00 | 35.33% | 0 / 0 / 0 |
| Sparse256 | 138.41 | 129.60 | 113.59 | 86.41 | 37.57% | 0 / 0 / 0 |
| Negative256 | 139.75 | 129.95 | 114.17 | 86.51 | 38.10% | 1 / 0 / 0 |
| Sparse384 | 121.61 | 120.85 | 108.34 | 81.50 | 32.98% | 302 / 81 / 17 |
| Weak256 | 256.00 | 256.00 | 254.22 | 256.00 | 0% | 512 / 471 / 512 |

The two primary paired gain lower bounds are 44.93 and 48.09 steps using the
frozen z=3.3 screen. Every null scenario has zero short-rate alerts in each
phase; each Wilson95 upper is 0.7447%, not a zero false-alarm guarantee.
There are no premature short-rate alarms in either phase.

The homogeneous confirmation mean-delay penalty is only 4.54 steps and now
passes the ten-step allowance. However, seven harmful discordances yield an
excess miss rate of 1.367%, exceeding 1%, and a simultaneous paired upper of
3.692%, exceeding 2%. The design phase also fails: ten short misses versus one
uniform miss, nine harmful discordances, and a 4.267% upper bound.

The short model misses all 512 weak changes, whereas the long model detects
41. The frozen relative guard compares to uniform, which also misses all 512,
so that cell passes only a weak relative screen. This is not adequate absolute
sensitivity or non-harm versus the long model. Late sensitivity improves but
still misses 17/512 changes (3.32%).

## Checks and cost

- Exact full replay passes in 13.70 seconds, including all recorded sources,
  per-stream alarms, observation counts and rate tapes.
- Focused race tests for short/long rates, augmentation and allocation pass in
  1.568 seconds. Package vet passes.
- The 32-window special case is identical to v76 across 400 updates. Tests
  check chronological ring wrapping against independent append-only histories,
  immutable state, unchanged acquisition/augmentation, sign and coordinate
  equivariance, invalid windows, ternary inputs, safe factors and grid optima.
- Apple M4, darwin/arm64, GOMAXPROCS10, three 500ms microbenchmark repetitions:
  balanced 471.3/470.5/465.1 ns per step; sustained positive
  783.6/785.9/775.9 ns per step. All report zero allocations.
- The benchmark includes preparation, validation, bounded evidence updates and
  local history updates. It excludes retrieval, random query execution, queues,
  network and persistence. It is neither an end-to-end nor worst-case bound.

The work remains bounded by four32-entry history scans, four8-entry rate scans,
two24-step derivative searches over12 outcomes, and eight signed starts. There
is no I/O or corpus-dependent loop in the new method.

## Next lead

The mean/tail tradeoff is now directly observed: faster rate adaptation reduces
delay while worsening some misses. Short-window estimator noise and prior
shrinkage are plausible contributors, not uniquely proven causes by this run.
Do not choose a compromise window from these outcomes and call it validated.

Next freeze a fixed/adaptive wealth mixture and test it on fresh data, retaining
the long model as a separate control. The fixed component can preserve evidence
when the learned rate is poorly estimated, but assigning it only part of the
initial wealth adds a real threshold penalty. Arithmetic mixture validity alone
does not prove non-harm, and no observed gain permits dropping weak or late
controls. Actual MMM integration and all seven research directions remain open.
