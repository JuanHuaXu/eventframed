# Conditional audit gate: identifiability component

## Result

The [frozen toy protocol](mmm-conditional-gate-v1-protocol.md) passes all four
finite screens. The context-conditioned audit detects a swapped conditional
outcome law on every one of 1,000 streams under each evidence schedule, with
zero flags on 1,000 stationary streams per schedule. The scalar-correctness
comparator flags none. This demonstrates a structural blind spot of a scalar
projection on the declared two-cell family; it does **not** validate an
operational Anti-Pigeon replacement or downstream forecast improvement.

The reference law is `(0.9,0.1)` and the live changed law `(0.1,0.9)` for
`P(Y=1|X)`. With uniform `X` and a frozen always-positive classifier, both
regimes yield i.i.d. scalar correctness with probability 0.5. Their mean
conditional total variation is 0.8. Thus no method receiving only that scalar
stream can distinguish these two regimes from the observations alone; more
scalar samples cannot fix the identifiability failure. This exact statement
depends on the declared equal context mix, fixed predictor and stationary
within-cell laws. It does not say the current daemon's full evidence stream is
always scalar or that every regime shift has this property.

## Evidence

Each arm has 1,000 independent reference/live streams and 512 live frames.
Both gates receive the same nominated, nonmissing labels at their recorded
arrival clocks. The reference uses 256 independent labels in each context.
The conditional rule uses finite-horizon Hoeffding radii with a union-bound
familywise budget of 0.02 and practical-equivalence width 0.10. No oracle
probability is an input to either monitor.

| Schedule | Regime | Scalar flags | Conditional flags | Mean usable live labels | Conditional clock p50 / p95 |
| --- | --- | ---: | ---: | ---: | ---: |
| Complete, immediate | Stable | 0/1,000 | 0/1,000 | 512 | N/A |
| Complete, immediate | Swapped | 0/1,000 | 1,000/1,000 | 512 | 32 / 46 |
| 25% audit, 20% missing, delay 0..31 | Stable | 0/1,000 | 0/1,000 | 99.298 | N/A |
| 25% audit, 20% missing, delay 0..31 | Swapped | 0/1,000 | 1,000/1,000 | 99.150 | 184 / 271 |

In the sparse delayed schedule, about 128 frames are nominated, 25-26 have
missing labels, and about three usable labels remain scheduled beyond clock
511 on average. The median conditional flag uses 34 arrived labels, compared
with 33 under complete immediate feedback. These are component detection
counts, not independent frame-level confidence intervals or a sample-size
claim for arbitrary conditional changes.

## Audit and limits

- [Raw 4,000-trial artifact](mmm-conditional-gate-v1.jsonl), SHA-256
  `b963bdc3fdb6667a89db181076dac145834714a15f21224a0f2222eb8a7fb986`.
- [Simulator](../../research/conditional-gate-v1.mjs), SHA-256
  `638b80d5416c76b224a88d55171a0f8e112c331f1169c10953640b66d343662e`.
- [Independent summary verifier](../../research/conditional-gate-v1-verify.mjs)
  checks every row's budget/accounting, uniqueness, source/protocol hashes,
  and the reported aggregates. A second generated artifact reproduces all
  4,000 trial rows and summary exactly apart from elapsed time.
- Negative controls reject future, missing, un-nominated and duplicate label
  delivery. Both monitor copies have identical observed counts and successes
  at every clock; predictions and labels are generated in separate roles.

The program completed the four cells in about 0.61 s on this host. That is
two million synthetic frames including random generation, audit checks and
output accumulation, **not** per-request service latency or a hardware-neutral
algorithmic benchmark. The conditional rule is constant work per arrived
label for two cells; scaling to many, adaptively created contexts needs its
own cap and simultaneous error budget.

The finite bound requires outcome-independent nomination, missingness and
arrival, independent reference/live labels within each cell, a fixed declared
cell family, and a trustworthy reference. It is not valid automatically for
informative feedback, source dependence, continuously changing contexts or
arbitrary target-law diameter. The test supplies no forecast update after a
split, so Goal 3's useful downstream result remains OPEN. No production,
whitepaper or remote changes were made.

## Next test

Freeze a simple post-authorization conditional forecaster before generating
new streams. Score each forecast before its own label can arrive, use only
arrived audit labels for updates, retain the scalar arm's baseline law, and
report stationary protection and shifted Brier at equal audit cost. This will
test whether restored identifiability translates into useful predictions in
this toy family, without upgrading the component certificate into general
Anti-Pigeon authority.
