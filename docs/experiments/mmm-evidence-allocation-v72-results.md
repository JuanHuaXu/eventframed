# Evidence allocation v72: validity rescue, performance target not reached

The corrected acquisition candidate **FAILS the full frozen screen**. It yields
a modest detection improvement and repairs a severe selection-induced false
alarm in the declared population-mean test. Neither result establishes a
production Anti-Pigeon policy or completes directions3/7.

## Selection changes the question being tested

In the heterogeneous boundary-null confirmation cell, the uniform target mean
is.15. Uniform sampling observes mean.151089 and produces0 alarms in512 streams.
Naive adaptive sampling observes mean.290199 and produces472/512 alarms
(92.19%). The identical adaptive observations, corrected by their known
selection probabilities, give mean.150303 and1/512 alarms (0.195%).

The corrected null alarm Wilson95 upper is1.098%, below the frozen2% finite
screen. The other three confirmation null cells have zero corrected alarms.
These are false alarms **under this population-mean null**. The heterogeneous
channels are not necessarily a safe abstraction under Anti-Pigeon's stronger
context-wise diameter criterion; no claim that they should share a posterior
follows from an average paired-correctness comparison.

This is the key integration lesson: choosing more revealing cases can change
their average behavior even without a change in the original population. A
sampling proposal or relevance score cannot silently become evidence for the
old target population. Inverse-probability correction works here because the
target masses and actual randomized propensities are known and bounded away
from zero. It does not repair unknown nomination probabilities, fabricated
provenance, or unobserved portions of the corpus.

## Detection and deadline results

10240 fresh paired streams, two phases, ten scenarios,512 streams per cell,
512 paid queries per arm per trajectory. Every learner sees one selected
channel per step; the simulator's other outcomes never enter its state.
Selected propensities are computed before the revealing outcome. All metrics
below are confirmation, with misses/premature alarms charged full remaining delay.

| Scenario | Uniform delay | Corrected delay | Gain | Uniform misses | Corrected misses |
| --- | ---: | ---: | ---: | ---: | ---: |
| Homogeneous128 | 156.29 | 157.43 | -0.73% | 0/512 | 0/512 |
| Sparse128 | 136.64 | 129.97 | 4.88% | 0/512 | 0/512 |
| Sparse256 | 137.91 | 131.06 | 4.97% | 0/512 | 0/512 |
| Negative256 | 138.43 | 132.56 | 4.24% | 1/512 | 0/512 |
| Sparse384 | 121.96 | 121.09 | 0.71% | 312/512 | 280/512 |
| Weak256 | 256.00 | 256.00 | 0% | 512/512 | 512/512 |

The two primary gains have positive paired z=3.3 lower bounds of3.81 and3.95
steps, but both fall short of10%. There are no premature confirmation alarms.
The weak-change case remains wholly undetected: a relative non-harm check does
not establish useful absolute sensitivity.

Sparse384 also fails the frozen conservative paired reliability bound. Although
the net miss rate improves by32/512 (6.25 percentage points), there are47 cases
where corrected sampling misses and uniform detects, versus79 improved cases.
The protocol upper-bounds excess misses using harmful discordances alone; its
simultaneous upper is13.053%, above2%. This is a failed certificate, **not
evidence of a higher overall miss rate**. A future predeclared paired-difference
bound can account for both directions of discordance; do not change this
study's frozen method after seeing the result. Even without that conservative
failure, the primary10% speed criteria remain unmet.

Naive targeting appears27-28% faster in the primary alternatives, but its
92.19% boundary-null alarm rate disqualifies that speed as a valid rescue.
Do not substitute its appealing delay numbers for the corrected candidate.

## Model and scope

The candidate estimates channel second moments from32 recent local outcomes,
uses a coverage-floored square-root proposal, and corrects D by .25/q. The
signed wealth factors remain positive and have null conditional expectation
at most1. The eight-start arithmetic mean therefore retains the declared
per-arm1% anytime bound in exact arithmetic under the stated sampling contract.

The adaptive-importance-sampling motivation is
[Ryu & Boyd, Adaptive Importance Sampling via Stochastic Convex Programming](https://web.stanford.edu/~boyd/papers/adaMC.html).
Our rolling discrete allocation is not their algorithm and inherits no
asymptotic optimality or convergence guarantee from it. Reduced importance
variance need not produce enough improvement in nonlinear stopping time.

This experiment isolates the acquisition/authority interface before wiring a
warning trigger. It is not a combined MMM observer, calibrated target-law
diameter certificate, real agent experiment, or test of causal identification.
Costs are equal paid paired-outcome queries; hardware/LLM acquisition prices
are not inferred from the simulation. Real retrieval would need explicit
target scope, randomized inclusion and trusted propensity recording.

## Verification and cost

- Frozen protocol: `mmm-evidence-allocation-v72-protocol.md`.
- Artifact: `mmm-evidence-allocation-v72.jsonl`, SHA256
  `1fce6db98b38b732519073509e12f10fa99d7dd59f9719ddfa1d4d977b5e61e0`.
-13 source/protocol/dependency hashes recorded and verified.
- Full10,240-stream deterministic replay passed in3.93s, including every
  propensity/channel/outcome tape hash, first alarm, query count and moment.
- Focused race checks passed in1.586s: selection-corrected expectation,
  factor positivity, proposal coverage, observation ordering/state isolation,
  invalid inputs, original pooled-gate parity and adjacent evidence tests.
- `go vet ./internal/observationgate` passed.
- Summary/reproduction: `node research/evidence-allocation-v72-summary.mjs`.
- Full replay: `EVENTFRAME_ALLOCATION_REPLAY=<artifact> go test ./internal/observationgate -run '^TestAllocationV72ArtifactReplay$' -count=1 -v`.

Isolated Apple M4, Go1.27.1 darwin/arm64, GOMAXPROCS10:

```text
BenchmarkAllocationObserve-10 2093984 282.5 ns/op 0 B/op 0 allocs/op
BenchmarkAllocationObserve-10 2123805 293.1 ns/op 0 B/op 0 allocs/op
BenchmarkAllocationObserve-10 2121866 282.0 ns/op 0 B/op 0 allocs/op
```

This bounded policy-plus-single-rate-gate loop excludes query execution,
random sampling, persistence, queues and full request latency. It is not a
speed comparison against v71's multi-rate/multi-arm research wrapper. No
production code path was enabled; no commits, pushes or whitepaper edits.

Next: preserve the validated selection-accounting invariant, investigate a
proper paired excess-risk bound and acquisition policies aligned with final
stopping time, and test a real warning-trigger contract before adoption. Do
not optimize the invalid unweighted arm or tune floors against confirmation.
