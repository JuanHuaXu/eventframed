# Predictable model-based bets v76 results

Overall: FAIL. The sparse-change speed requirement passes for the first time
in this sequence, but homogeneous-change protection fails in both phases.
Do not enable the candidate in production or count a research direction complete.

## Frozen experiment

[Protocol](mmm-predictive-bet-v76-protocol.md),
[raw artifact](mmm-predictive-bet-v76.jsonl), and
[evaluator](../../research/predictive-v76-summary.mjs) describe 10,240 fresh
paired streams: two phases, ten scenarios, 512 streams per cell, 512 steps.
The confirmation phase was not used to modify the candidate. The artifact has
19 source/evaluator hashes; SHA256:
`061aea3abec071d58fee936edd22cf63de5a873ea34c40835f49dcdc49a29716`.

The method selects a sign-specific betting rate from past ternary observations,
before drawing the next query. It uses the same observation allocation and
augmentation as v74. This is a research adaptation of predictable betting, not
a reproduction of [Waudby-Smith and Ramdas](https://arxiv.org/abs/2010.09686).
Known actual sampling probabilities, predictable rates and bounded factors
preserve the conditional-mean null argument. They do not guarantee power or
validate an Anti-Pigeon target-law diameter certificate.

## Confirmation results

Restricted delay counts an undetected change at the remaining horizon. Each
row contains 512 trajectories. Negative gain means slower than uniform.

| Scenario | Uniform delay | Fixed augmented delay | Predictive delay | Gain vs uniform | Uniform / predictive misses | Result |
| --- | ---: | ---: | ---: | ---: | ---: | --- |
| Homogeneous, step 128 | 152.75 | 153.66 | 180.59 | -18.23% | 0 / 2 | FAIL |
| Sparse, step 128 | 139.69 | 129.23 | 117.51 | 15.88% | 0 / 0 | PASS finite screen |
| Sparse, step 256 | 139.45 | 130.87 | 115.15 | 17.42% | 2 / 0 | PASS finite screen |
| Negative, step 256 | 137.25 | 129.34 | 114.00 | 16.94% | 0 / 0 | PASS finite screen |
| Sparse, step 384 | 120.91 | 120.13 | 107.80 | 10.84% | 306 / 84 | PASS relative screen |
| Weak, step 256 | 256.00 | 256.00 | 254.23 | 0.69% | 512 / 480 | PASS relative screen only |

Primary paired gain lower bounds are 18.53 and 20.52 steps, using the frozen
z=3.3 screen. All four confirmation null scenarios have zero alerts; each
Wilson 95% upper bound is 0.7447%, not a zero false-alarm probability.
There are no premature confirmation alerts.

Homogeneous detection is 27.84 steps slower, beyond the ten-step allowance.
Its simultaneous paired excess-miss upper bound is 2.098%, beyond 2%.
The design phase also fails: 15.80% slower, three misses versus none, and
2.446% paired upper. Thus the failure is not confined to one confirmation cell.
The paired bounds use the v73 procedure and alpha=.05/12, as frozen before data.

Late misses improve materially but remain 84/512 (16.41%). Weak changes remain
mostly undetected: 480/512 (93.75%). Passing non-harm relative to a weak control
does not establish adequate absolute sensitivity. Detected-only weak delay is
227.72 steps over only 32 detections, not all 512 trajectories.

## Verification and cost

- Exact full replay of all 10,240 records passes in 10.09 seconds, including
  source hashes, rate tapes, old-control parity and query counts.
- Focused race tests for predictive bets, augmentation, allocation and start
  pooling pass in 1.585 seconds. Package vet passed before the experiment.
- Optimizer tests compare predicted growth with a dense rate grid, check
  all-channel endpoint bounds, invalid-input isolation and immutable snapshots.
- Three Apple M4 microbenchmarks: 514.5, 524.8 and 520.8 ns/op, zero allocations.
  This cyclic balanced-input fixture mostly selects zero rates. It includes
  bounded policy/gate updates but does not characterize sustained bisection,
  worst-case latency, retrieval, persistence, queues or end-to-end serving.

## Interpretation and next lead

The v75 growth diagnostic translates into measured sparse detection gains,
but not uniform gains. A plausible cause of homogeneous harm is the lag and
estimation error of the learned rate after a change; this is a hypothesis,
not established by the aggregate results. The fixed augmented control does
not show comparable homogeneous harm on these same streams.

Next isolate that mechanism and test a predeclared fixed/adaptive wealth
mixture or bounded adaptation safeguard on new streams. Mixture validity does
not imply no delay penalty: dividing initial wealth can itself slow detection.
Retain this failure and the original speed/non-harm requirements. No additional
threshold tuning, live MMM integration, production change or completed research
direction is justified by this result.
