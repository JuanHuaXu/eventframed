# Start-pooled gate v71: partial improvement, full screen failed

Direction3 has a new mathematically justified partial improvement, not a passed
replacement. All three candidates (old grid, pooled fixed, pooled grid) fail the
full frozen speed/reliability screen. Production and MMM sharing stay unchanged.

## Fresh paired evidence

10240 independent synthetic streams across design/confirmation, ten scenarios,
512 streams per cell,512 steps each. Four gates see exactly the same outcome
tape. Full tapes and first alarms are retained. The two old controls match the
existing implementation at every step. On all tapes, the first pooled alarm is
no later than its corresponding old gate's first alarm, when that old gate fires.

Confirmation restricted mean delays (misses and premature alarms charged the
full remaining horizon) and deadline misses:

| Gate | Moderate128 delay | Misses/512 | Moderate256 delay | Misses/512 | Negative256 misses/512 |
| --- | ---: | ---: | ---: | ---: | ---: |
| Fixed | 157.91 | 0 | 155.23 | 18 | 13 |
| Grid | 139.86 | 0 | 137.02 | 29 | 25 |
| Pooled fixed | 154.46 | 0 | 151.42 | 15 | 9 |
| Pooled grid | 137.80 | 0 | 135.15 | 25 | 22 |

Pooled fixed improves the two primary delays by2.19% and2.46%, below the frozen
10% requirement. Paired z=3.3 lower bounds on gains are positive:2.54 and2.68
steps. It passes the other confirmation checks, with zero harmful miss
discordances in every alternative. This is not zero population uncertainty:
the simultaneous one-sided upper bound is1.242% for0/512.

Pooled grid improves primary mean delay by12.74% and12.94% but still adds7/512
late moderate misses (1.367%) and9/512 negative-shift misses (1.758%). The
corresponding harmful-discordance upper bounds are3.640% and4.211%, exceeding2%.
Thus pooling improves the existing grid's deadline misses but does not fully
rescue it against the fixed reference. The old grid fails these limits too.

All confirmation null checks pass. In the high-variance null boundary scenario,
fixed/grid/pooled-fixed/pooled-grid have2/2/3/2 alarms in512 streams; the largest
Wilson95 upper bound is1.708%. The other four null cells have zero alarms per
arm. There are no premature alarms in the confirmation alternatives. Pooled
fixed therefore is not falsely described as having identical observed false
alarms to fixed: there was one additional boundary-null alarm.

Weak-shift detection remains poor: fixed/grid/pooled-fixed/pooled-grid miss
511/502/510/499 of512. Passing a relative guardrail is not adequate absolute
weak-change sensitivity. Detected-only delays remain in the raw summary and
are not substituted for restricted delay.

## What the mathematics establishes

Each fixed-start signed evidence process retains v1's conditional null. The
arithmetic average of all eight start processes, including inactive unit wealth,
is a nonnegative supermartingale starting at1. Threshold100 gives the same
per-arm theoretical1% anytime false-alarm bound as the old union rule's800
threshold per start. Overlapping windows do not need independence for averaging.
This is evidence-wealth averaging, not posterior sharing or multiplying copies
of an observation as independent evidence.

Whenever any start reaches800, its contribution alone makes the average at
least100. Hence first-alarm dominance holds in exact arithmetic. An earlier
premature alarm can still harm post-change detection, which is why that outcome
is kept as a failure in both protocol and scorer. Finite simulations test the
implementation; they do not establish that real-world evidence satisfies the null.

The construction adapts arithmetic evidence merging from
[Vovk & Wang (2021)](https://arxiv.org/abs/1912.06116), with the time-uniform
framework of [Howard et al. (2021)](https://arxiv.org/abs/1810.08240).
Each arm has its own bound; running/selecting all arms does not automatically
retain a joint1% error budget.

## Checks and reproducibility

- Frozen protocol: `mmm-start-pool-v71-protocol.md`.
- Raw artifact: `mmm-start-pool-v71.json.gz`, SHA256
  `a45070726fdfe74e55c63a174f2f944acbaaebd0a679dc4901eb79ad013e2bba`.
- Experiment completed in17.95s with all10240 tapes, old-control parity and
  pathwise first-alarm dominance checked during generation.
- `node research/start-pool-v71-summary.mjs` independently verifies11 hashes,
  row/tape completeness, dominance, raw-derived summary counts/delays, exact
  binomial discordance bounds and the all-false final pass vector.
- Focused race tests passed in1.549s: direct ordinary wealth arithmetic versus
  log-space pooling, sign symmetry, invalid-input state preservation, old
  original-gate parity, predictable updates, null factors and start/latch behavior.
- Binomial inversion matches analytic zero-count and n=2,k=1 solutions and
  passes monotonicity checks. Its simultaneous bound over30 declared comparisons
  avoids the earlier degenerate paired-normal rare-event interval.
- `go vet ./internal/observationgate` passed.

Rerun the experiment with an unused output path:
`EVENTFRAME_START_POOL_ARTIFACT=<new-path> go test ./internal/observationgate -run '^TestStartPoolV71Experiment$' -count=1 -v`.
No parameter fitting occurred between design and confirmation.

## Cost and next lead

Isolated Apple M4, Go1.27.1 darwin/arm64, GOMAXPROCS10, three500ms benchmark
replicates after race tests finished:

```text
BenchmarkStartPool/fixed_grid_control-10 501097 1196 ns/op 0 B/op 0 allocs/op
BenchmarkStartPool/fixed_grid_control-10 502240 1189 ns/op 0 B/op 0 allocs/op
BenchmarkStartPool/fixed_grid_control-10 511058 1200 ns/op 0 B/op 0 allocs/op
BenchmarkStartPool/start_pool-10         311967 1929 ns/op 0 B/op 0 allocs/op
BenchmarkStartPool/start_pool-10         314452 1922 ns/op 0 B/op 0 allocs/op
BenchmarkStartPool/start_pool-10         319960 1934 ns/op 0 B/op 0 allocs/op
```

The multi-arm research wrapper adds about0.73us per observation (roughly61%),
with no allocations. This is a gate microbenchmark, not full serving latency or
an optimized single-arm implementation. No production performance change occurs.

Do not weaken10% or the miss limits to relabel this as a pass. The proof-backed
first-alarm improvement is retained as a candidate component. A useful next
direction is investigation-trigger integration: use earlier disagreement to
request discriminating evidence while keeping final sharing authority separate,
and test total acquisition cost, final split timing and false revocation jointly.
That may connect directions3/7, but remains untested and cannot inherit this
isolated screen's guarantees without a valid predictable evidence contract.
