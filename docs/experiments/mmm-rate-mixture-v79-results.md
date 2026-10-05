# Fixed/long/short evidence mixture v79 results

PASS of the frozen finite gate-level screen in both phases. This is not a
completed MMM integration, general reliability claim or production approval.
Weak sensitivity and member-specific sharing validity remain unresolved.

[Protocol](mmm-rate-mixture-v79-protocol.md),
[raw artifact](mmm-rate-mixture-v79.jsonl),
[evaluator](../../research/rate-mixture-v79-summary.mjs),
[implementation](../../internal/observationgate/rate_mixture.go).

10,240 fresh paired streams, 24 source/evaluator hashes, no tuning between phases.
Artifact SHA256:
`1b1d042d7eadf1f5d616ccd456368e045ad49d22683ead4eee38e81b1fa994a0`.
Weights remain (.5 fixed, .25 long, .25 short), frozen before data. The output
is a weighted sum of current evidence wealth, not independent multiplication
or a vote over latched alarms. Each component continues updating after an alarm.

## Confirmation results

512 trajectories per scenario. Restricted delay counts misses at the remaining
horizon. All controls are from this same fresh paired experiment.

| Scenario | Uniform delay | Fixed augmented | Long rate | Short rate | Mixture | Mixture gain vs uniform | Uniform / mixture misses |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Homogeneous128 | 153.41 | 154.83 | 177.94 | 157.91 | 154.25 | -0.54% | 0 / 1 |
| Sparse128 | 140.86 | 129.95 | 118.04 | 92.66 | 103.45 | 26.56% | 0 / 0 |
| Sparse256 | 138.44 | 130.05 | 113.54 | 86.05 | 97.99 | 29.22% | 2 / 0 |
| Negative256 | 140.90 | 130.40 | 114.47 | 86.71 | 98.60 | 30.03% | 0 / 0 |
| Sparse384 | 121.62 | 120.48 | 108.28 | 82.09 | 93.30 | 23.29% | 320 / 58 |
| Weak256 | 256.00 | 256.00 | 254.37 | 256.00 | 255.79 | 0.08% | 512 / 507 |

Primary paired gain lower bounds are 33.65 and 36.78 steps using the frozen
z=3.3 rule. The design primary gains are 25.80% and 28.00%; its protection
cells also pass. No mixture premature alarms occur in either phase.

All four null scenarios have zero mixture alarms in each phase, with individual
Wilson95 upper bounds 0.7447%. The homogeneous confirmation cell has one harmful
discordance, observed excess misses 0.1953%, and simultaneous paired upper
1.7189%, below the frozen 2% ceiling. Its 0.832-step delay harm is below the
ten-step allowance. These are finite-sample screens, not a zero-risk result.

## What was rescued and what was not

The mixture preserves substantial sparse detection gains while meeting the
original uniform-control mean-delay and excess-miss requirements. Relative to
the standalone short model, it reduces homogeneous confirmation misses from
four to one. It is not faster than every component in every case.

There is a real cost to distributing initial wealth: late misses are58/512
versus32/512 for short alone, and weak detections are5/512 versus37/512 for long
alone. Weak misses remain99.02%. Thus a passing uniform-relative screen does
not establish adequate absolute sensitivity or dominance over each component.
No prior v76/v78 failure is reclassified by this separate experiment.

The conditional-null proof applies to the declared bounded population-mean
statistic, with actual sampling probabilities and predictable corrections.
It does not certify target-law diameter or each member's equivalence. Opposing
member effects can cancel in a mean. Absence of an alarm must never be used as
permission to merge or continue sharing without the separate Anti-Pigeon
member-specific contract. The heterogeneous and cancelling null controls make
this distinction particularly important.

## Verification and cost

- Full replay of all10,240 records, source hashes, component alarms and query
  counts passes in25.18 seconds. Component tapes match the archived algorithms
  run on the same fresh seeds.
- Direct probability-space evidence matches log-space mixture calculation
  through512 updates. Initial wealth is1; no independence is assumed.
- Invalid evidence, mismatched snapshots and an invalid last component leave
  all state unchanged. Updates commit together using a bounded value copy.
- Focused race tests pass in1.613 seconds; package vet passes.
- Apple M4, darwin/arm64, GOMAXPROCS10, three500ms microbenchmarks:
  balanced1900/1897/1905 ns per update; positive2059/2053/2067 ns.
  Both report zero allocations. Two bounded rate preparations, three evidence
  updates and their weighted pooling are included; retrieval, query execution,
  queueing, persistence and network are excluded. This is not serving latency.

## Next boundary

Preserve this candidate without retuning its weights. Test its investigation
trigger in the research MMM member-split workflow, including cancellation,
delayed/missing evidence, independently generated changes and equal acquisition
cost. Keep investigation separate from sharing authority. An aggregate trigger
cannot replace the per-member validity test; route or supplement it accordingly.
Carry absolute weak/late detection limitations into that integration rather than
calling relative non-harm sufficient. All seven roadmap directions remain open.
