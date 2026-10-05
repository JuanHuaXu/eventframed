# Event-count acquisition proxy: no rescue

## Verdict

**FAIL; not adopted.** The frozen acquisition-only substitution passes 2/4
required reverse-recovery screens and 14/16 nonharm screens. The preceding
coupled observer passed 3/4 and 14/16, respectively. Neither immediate forward
failure is cleared. These are consumed-data exploratory screens, not fresh
confirmation or simultaneous error guarantees. All seven research goals remain
open.

The intervention replaces event-count guide dispatch with the existing
event-subset observer when that model is available. The actual count forecast
remains in the scored mixture. Model fitting, priors, weight updates, split
rules, and acceptance thresholds are unchanged. Missing subset models retain
the old behavior. See [frozen contract](mmm-window-proxy-v1-contract.md).

## Forecast quality

Post-change requested-view Brier (lower is better):

| Cohort | Change / schedule | Original | Previous coupled | Proxy |
| --- | --- | ---: | ---: | ---: |
| 1 | Majority to parity / immediate | .243420 | .251267 | .250049 |
| 1 | Majority to parity / delayed | .273706 | .274734 | .274038 |
| 1 | Parity to majority / immediate | .231338 | .218184 | .218037 |
| 1 | Parity to majority / delayed | .259748 | .250972 | .251089 |
| 2 | Majority to parity / immediate | .240403 | .249508 | .253066 |
| 2 | Majority to parity / delayed | .270235 | .270248 | .270416 |
| 2 | Parity to majority / immediate | .218989 | .207545 | .207879 |
| 2 | Parity to majority / delayed | .252457 | .246148 | .246150 |

The two cohort-1 reverse cases pass. Cohort-2 immediate reverse has mean gain
.011110 but its paired mean +/- 3.5SE interval is [-.000092, .022312]; delayed
reverse has gain .006307 and interval [-.001057, .013671]. Neither passes the
positive-lower-bound rule. Both immediate forward cases fail nonharm.

Compared with the fixed-acquisition arm, total unique-coordinate cost increases
on 101 schedules, is equal on 121, and decreases on 34. Compared with the previous
coupled arm, the counts are 28, 186, and 42. These are schedule counts, not mean
percentage cost changes. Shared monitoring reads are charged only once.

The count guide's lack of uniform-input joint coherence is a mathematical
limitation, but this ablation does not establish it as the cause of the quality
regression. A coherent replacement proxy alone is insufficient here.

## Verification

- All 256 schedule-runs complete. Disabled intervention reconstructs the prior
  coupled records exactly; enabled runs preserve the fixed arm exactly.
- Native collection: 41.87 seconds; full native replay: 41.90 seconds. Whole raw
  artifacts are byte-identical (`cmp` exit 0).
- Independent scorer verifies 895 captured source hashes and 5,111,808 scalar
  checks; maximum numerical discrepancy is 1.1102230246251565e-15. Replay summary
  is byte-identical to the original summary.
- Focused race contracts and `go vet ./internal/observationgate` pass. This is
  not a full race-instrumented experimental collection.

Raw/replay SHA256:
`cdd7810b488215cb6556e3912b2490bd8c11dbd69aac61267b311e2498d0be7b`.

Artifacts: [raw](mmm-window-proxy-v1.jsonl),
[replay](mmm-window-proxy-v1-replay.jsonl),
[summary](mmm-window-proxy-v1-summary.json),
[replay summary](mmm-window-proxy-v1-replay-summary.json).

## Performance

Apple M4, darwin/arm64, three sequential benchmark repeats:

| Focused prediction fixture | ns/op | B/op | Allocations/op |
| --- | ---: | ---: | ---: |
| Existing count guide | 11066-11134 | 11405 | 39 |
| Subset proxy | 4969-4981 | 5635 | 28 |

Command: `go test ./internal/observationgate -run '^$' -bench '^BenchmarkWindowProxyPredict$' -benchmem -count=3`.
This fixture forces the affected guide branch; it excludes fitting, storage,
retrieval, queueing, and service tails. A faster failed forecast candidate is not
a deployment improvement.

## Next investigation

Historical family-model substitution is not an untested rescue: v91 fails
sparse-sample protection, and the same-window family-evidence experiment fails
its primary recovery screens (0/16 gains, 149/168 nonharm). Do not repeat it
without distinguishing the mechanism from those failures.

Next isolate fresh-evidence sufficiency from acquisition and mixture learning
on these same consumed trajectories. A hindsight change boundary or full-input
oracle may define diagnostic headroom only; neither may enter a deployable
policy. Preserve the original forecast control and both shift directions. Only
after locating a recoverable bottleneck should another candidate use fresh
confirmation data. No production, whitepaper, commit, or remote change.
