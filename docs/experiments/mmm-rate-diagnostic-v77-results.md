# Rate adaptation diagnostic v77 results

Finding: slow adaptation of the rate model is a supported mechanism behind
v76's homogeneous-change failure. This is post-hoc analysis, not a successful
rescue or fresh confirmation. The v76 FAIL classification remains unchanged.

[Protocol](mmm-rate-diagnostic-v77-protocol.md),
[raw records](mmm-rate-diagnostic-v77.jsonl),
[diagnostic code](../../internal/observationgate/rate_diagnostic_test.go), and
[descriptive summarizer](../../research/rate-diagnostic-v77-summary.mjs).

All 10,240 original streams are retained. The actual candidate's rate tape
hashes and actual/fixed first alarms match v76 exactly. The oracle uses the
simulator's current channel law only to select rates; it retains the candidate's
query policy, augmentation and eight evidence starts. It knows the change time
and is not implementable evidence or a candidate for deployment.

Artifact SHA256:
`9f1bcdbbd29e883b24fa60c1e010c0df4a3b9b3f3128ae3e6b36e4ea7e097c9f`.

## Mechanism evidence

For the homogeneous confirmation scenario (512 trajectories), conditional
expected log-factor in the favorable sign is:

| Steps relative to change | Fixed rate | Actual learned rate | Oracle rate | Actual zero-rate fraction |
| --- | ---: | ---: | ---: | ---: |
| Last 64 before change | -0.04185 | 0 | 0 | 100% |
| 0-31 | 0.04686 | 0.000075 | 0.05999 | 99.43% |
| 32-63 | 0.04684 | 0.01047 | 0.05998 | 58.61% |
| 64-127 | 0.04679 | 0.04655 | 0.05971 | 2.91% |
| 128-255 | 0.04680 | 0.05528 | 0.05954 | 0% |
| 256-383 | 0.04681 | 0.05511 | 0.05958 | 0% |

These are exact conditional expectations under each simulator law, averaged
over realized past histories, not realized log increments or Brier scores.
The favorable rate averages only 0.00033 in the first post-change window,
versus an oracle 0.44694. By steps 128-255 it reaches 0.42738 versus 0.44206.
The oracle-growth gap falls from 0.05992 to 0.00426 per observation.

The rate estimator retains 32 observations per channel. At near-uniform
acquisition that spans approximately 128 global observations, so the first
32-64 post-change observations still coexist with substantial old evidence.
The traces directly show near-zero betting during this interval. Later
estimation loss persists, but is substantially smaller. This supports an
adaptation-lag explanation rather than a claim of no available rate headroom.

Avoiding bets before the change saves negative expected growth; that benefit
does not rescue the subsequent homogeneous delay. Aggregates do not uniquely
partition stopping-time delay into pre-change wealth, model error and start
mixture effects. No hindsight reset was introduced.

## Counterfactual delays

Same confirmation observations, 512 trajectories per row. Restricted delays
charge misses at the remaining horizon, not at zero.

| Scenario | Fixed augmented | Actual v76 | Oracle rates | Actual / oracle misses |
| --- | ---: | ---: | ---: | ---: |
| Homogeneous128 | 153.66 | 180.59 | 101.56 | 2 / 0 |
| Sparse128 | 129.23 | 117.51 | 56.73 | 0 / 0 |
| Sparse256 | 130.87 | 115.15 | 53.97 | 0 / 0 |
| Negative256 | 129.34 | 114.00 | 54.04 | 0 / 0 |
| Sparse384 | 120.13 | 107.80 | 50.68 | 84 / 4 |
| Weak256 | 256.00 | 254.23 | 207.38 | 480 / 151 |

The oracle is diagnostic, not a universal bound on stopping time: one-step
expected-log optimization does not solve the full stopping-time problem.
Even this oracle misses 151/512 weak changes. The weak-signal problem therefore
cannot be declared solved by eliminating the observed rate lag alone.

## Verification and next experiment

- Independent 60-point generator enumeration matches explicit probabilities
  in all ten scenarios before and after the change.
- Oracle optimization passes dense-grid checks and all-channel factor bounds.
- Contract race tests pass (1.314 seconds); package vet passes.
- Exact full artifact replay passes (11.674 seconds), including the original
  source checks and all original candidate tape/control comparisons.
- The descriptive summarizer verifies source hashes, seeds and window counts.

Next test a faster, separate rate estimator while keeping acquisition and
augmentation unchanged, so the intervention addresses the identified lag.
Shorter windows may increase rate noise; use fresh streams, the unchanged
null/speed/non-harm gates and all weak/late negative controls. A fixed/adaptive
wealth mixture remains another lead but has a real initial-wealth penalty.
Do not use the oracle's rates, generator identities or change times in either
candidate. No production changes, real-agent validation or completed research
direction follow from this diagnostic.
