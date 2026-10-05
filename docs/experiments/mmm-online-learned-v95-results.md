# v95: online correction with learned experts

## Verdict: FAIL, 54/58 primary gates

All48 full-stream non-harm gates and all6 stationary interaction-gain gates
pass. All4 recovery gates fail. Do not substitute cumulative protection or
stationary accuracy for recent-regime safety. No complete research direction
is finished; no production path was changed.

The [frozen protocol](mmm-online-learned-v95-protocol.md) evaluates768 new
streams, two phases with disjoint rule pools,32 paired streams per case,
12 cases,16 initial training labels and256 scored steps. Each stream includes
five arms and two fixed observation views. Full-frame training evidence is
shared by all arms and both views. Model refits occur every32 steps using the
latest64 received frames; online weights persist through refits. Every current
outcome is sampled after all forecasts exist. No change-point oracle enters
learning or weight updates.

See [raw records](mmm-online-learned-v95.json),
[independent summary and all gates](mmm-online-learned-v95-summary.json), and
[evaluator](../../research/online-learned-v95-summary.mjs).
Intervals below use the frozen paired trajectory z=3.5 normal approximation;
they are not anytime confidence sequences or exact finite-sample coverage.

## Confirmation results

Full-view all-stream expected Brier, generic to primary online-share:

| Case | Generic | Online | Interpretation |
|---|---:|---:|---|
| Parity3 | 0.076755 | 0.069238 | Gain gate passes |
| Parity4 | 0.092703 | 0.072571 | Gain gate passes |
| Complement4 | 0.097576 | 0.072675 | Gain gate passes |
| Majority3 | 0.076064 | 0.076241 | Small harm within declared margin |
| Multiplexer3 | 0.074629 | 0.074843 | Small harm within declared margin |
| Null | 0.267851 | 0.264651 | Better Brier, accuracy remains50% |

Parity4 all-stream expected accuracy improves88.00% to91.68%. Late-half
accuracy improves93.31% to94.96%, with Brier0.068154 to0.051103. These are
simulator expected accuracies including5% label noise, not agent-answer rates.
They are not a reproduction of the original MMM94.7% experiment.
Mask63 parity4 accuracy is only62.88%; full-view results do not solve partial
observation. Its mean Brier gain0.006683 has interval[-0.000237,0.013604].

Recovery must be reported separately:

| Late-half transition | Generic Brier | Online Brier | Gain interval | Verdict |
|---|---:|---:|---|---|
| Majority to parity | 0.204243 | 0.201429 | [-0.000459,0.006087] | Insufficient gain |
| Parity to majority | 0.172585 | 0.184354 | [-0.016488,-0.007050] | Harm |

The design phase also fails both recovery gates. Whole-stream parity-to-majority
Brier nevertheless improves0.147212 to0.140014 in confirmation, illustrating
how earlier gains conceal late harm. Skeptical BMA is not a demonstrated rescue:
its confirmation late parity-to-majority Brier is worse still,0.190748.
The standalone parity expert has0.264845 in that segment. Distinguish selector
inertia, stale model windows and family mismatch before combining more changes.

## Verification and performance

- Generation:92.367s package time, all prefix bounds pass with zero recorded defect.
- Exact768-record replay:93.743s package time; all source hashes and records match.
- Race smoke:2.800s; paired deterministic replays and prefix checks pass.
- Vet, independent summary reproduction and diff whitespace checks pass.
- Artifact SHA256: eebf1f49311b5009f710a78590bf5969f9ee4d64ac19b57dd0d257ae9210ff36.

Apple M4,Go1.27.1 darwin/arm64,GOMAXPROCS10: three single-stream benchmark
repetitions take119.09-145.54ms, allocate13.42-13.45MB and950-1078 objects.
The entire fixture includes eight refit rounds, three separately constructed
models (including duplicated work for the BMA control), all five arms, two
views, RNG, scoring and tape hashing. It is neither candidate-only cost nor
per-query latency. The first repetition's higher time is retained, not removed.
See [raw timing](mmm-online-learned-v95-benchmarks.txt). v94's~114ns arithmetic
cycle remains a distinct boundary and must not be called the cost of learning.

## Next lead

Test [interval-aware online correction](../../research/interval-brier-proposal.md)
on fresh streams, first over identical expert forecasts to isolate selector
history. If stale forecasts remain the limiting factor, add shorter model
windows in a separate, fully costed experiment. Strongly adaptive online
learning motivates interval comparisons; its asymptotic theorem does not
establish our finite1% or recovery criteria. Delayed and selective feedback,
real-agent validation and durable concurrent serving remain open.
