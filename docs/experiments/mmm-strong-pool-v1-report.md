# Broader forecast pool and exact incremental computation

Status: quality FAIL as a broad rescue; exact-computation speedup verified.
All seven research goals remain open. The consumed independent-v1 cohort has
672 trajectories and eight indices per comparison cell; it is not fresh
confirmation. See the [frozen contract](mmm-strong-pool-v1-contract.md).

## Quality

The full pool exposes Markov12 and the already-computed raw generic64/32,
Boolean64/32 and segment64 forecasts. Half the prior mass stays on Markov;
the alternatives split the other half equally. The no-segment control removes
only segment64 and redistributes alternative mass uniformly. Strong and linear
substitutions use identical eta2 weights and the same delayed sharing schedule.

| Arm | Whole Brier | Terminal64 | Protection / 672 | Gains / 96 | Harm windows / 5376 |
| --- | ---: | ---: | ---: | ---: | ---: |
| Full pool, strong | .154982984 | .143918839 | 625 | 1 | 59 |
| No segment, strong | .155394439 | .144754403 | 625 | 1 | 88 |
| Full pool, linear | .154884377 | .143874251 | 626 | 1 | 48 |
| No segment, linear | .155282652 | .144722106 | 626 | 1 | 83 |

Markov whole Brier is .157516545. Full strong improves that pooled mean by
about 1.61% relative. Stationary Brier improves from .119723242 to .118438460;
changing Brier improves from .218930662 to .214367835. Full strong is also
numerically better than its no-segment control overall and at terminal64.
These numerical comparisons do not establish simultaneous superiority.

The earlier two-head strong arm had 165 harmful windows and 608 protection
passes; exposing raw heads reduces those harms and increases the pass count.
However, qualifying recovery comparisons fall from seven to one. This tradeoff
is not a completed recovery rescue. All original controls and failures remain.
Intervals retain the original +/-3.5SE form on the eight available indices,
with .01 upper non-harm and .005 mean/positive-lower gain criteria. This is not
the original32-index study, a confidence sequence, or a population certificate.

## A fixed-cohort feasibility boundary

`research/recovery-feasibility.mjs` removes ALL forecast-family and guard
restrictions. Expected binary Brier is `(p-Q)^2 + Q*(1-Q)`, minimized at p=Q.
The greatest possible mean gain over a fixed control is therefore the mean
of `(control-Q)^2`. Q enters this evaluator only, never the candidate.

Of the 96 recovery comparisons, one cannot reach .005 even with oracle Q:
phase0/case19/immediate terminal64 against Markov has maximum gain
.001448211815. It is mathematically impossible to pass all96 on these exact
records. The result and replay are saved, and no gate was removed or changed.

This does NOT establish impossibility on the original32-index cohort or the
external population. The other95 measured comparisons have adequate oracle
headroom; the candidate still fails most of them. Consequently this one
unattainable gate does not explain away the broader learning shortfall. Stop
searching for all96 passes on this fixed cohort; first check the analogous
ceiling on the original archived protocol before changing evaluation design.
No sample exclusion or outcome-driven threshold repair is authorized here.

## Component verification

The multi-head kernel passes 582 checks covering independent latent-path
enumeration (maximum discrepancy 3.33e-16), two-head equivalence, permutation
and outcome symmetry, static immediate-feedback bounds, and as-of poisoning.
Full screen scores match original source controls, and the quality replay is
byte-identical. No immediate-feedback theorem is inherited by the delayed,
missing-label, sharing adaptation.

## Exact incremental rescue for computation

Profiling the reference scope shows repeated reconstruction of the unchanged
prefix. The replacement caches prefix states and replays from the earliest
newly arrived origin. Late feedback revises internal messages, never already
issued forecasts. Zero-delay labels are applied at the next issue, not to
their own prediction. Simultaneous arrivals are incorporated together.

`research/incremental-pool.test.mjs` verifies 8,192 forecast states bit-for-bit
against the reference, including no sharing, sharing, immediate feedback,
missing labels, delayed labels and simultaneous arrivals. Future/unavailable
label poisoning does not alter previous outputs. Full-data verification then
checks all172,032 strong forecasts bit-identically against the saved screen.

Three warm weights-only timing rounds, with alternating method order:

| Feedback schedule | Full refilter, us/forecast | Incremental, us/forecast |
| --- | ---: | ---: |
| Immediate | 19.09-19.24 | .258-.264 |
| Delayed/missing | 15.27-15.40 | .841-.852 |

This is approximately 73x and 18x faster respectively on this workload.
Recomputed transitions fall from22,106,112 to1,030,484. Timing excludes
aggregation, I/O, source preparation, fitting, scoring, persistence, queues
and concurrent serving. It is a research-kernel improvement, not a daemon
latency or loaded-learning success claim. The original source collection's
545.32s cost remains charged.

For T frames, K heads and bounded maximum delivery lag D, recomputation is
O(T*K*(D+1)); immediate feedback still revisits the previous origin once.
Unbounded delays retain O(T^2*K) worst-case time. This batch implementation
keeps O(T*K) state, not constant memory for an unbounded daemon. Ring-buffer
expiry, persistence and concurrent ownership are unimplemented here.

## Artifacts

- `mmm-strong-pool-v1-{screen,replay,timing,replay-timing}.json`
- `mmm-recovery-feasibility-v1{,-replay}.json`
- `mmm-incremental-pool-v1-benchmark.json`
- Kernels/tests: `research/strong-brier-pool*` and `research/incremental-pool*`

SHA256:

```text
incremental kernel 899d864b48c37fa396e14205c36fed1e6928b49d49d4feed74b540884190bae3
quality screen 3deb26769904241ed4835612790921e52170b6d70018455a588af3d43866aa3e
benchmark 4ff4410874089660a89bfe9bef0d9b6d4f8cb212764c415e9ff3242546396a5f
```

All commands refuse existing output paths. No Go runtime, production, private
data, whitepaper, commits, or remote repositories were changed.
