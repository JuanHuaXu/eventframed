# Arrival-learning dependence breadth

## Result

Partial exploratory support, not confirmation or adoption: 19/20 changed-case
gain screens and 48/48 full/post nonharm screens passed. All seven research
goals remain open. The method and thresholds were not changed during collection.

The unchanged four-arm innerArrivalRun driver compared original learning,
outer-only arrival learning, full outer/inner arrival learning, and fixed-view
outer arrival learning. Twelve dependence/noise/shift cases, two consumed
cohorts of 16 trajectories, and immediate/delayed schedules produced 768
schedule-runs of 512 frames. Delay is jitter 0..31 with missing probability .2.
The raw phase name `confirmation` means the second cohort here, NOT untouched
confirmation: both seed allocations had already been used.

## Second-Cohort Delayed Results

Post-window Brier (lower is better), original versus full arrival learning:

| Case | Original | Full | Gain screen |
| --- | ---: | ---: | --- |
| noise10_bit | .280019 | .265025 | Pass |
| noise10_xor2 | .284749 | .264137 | Pass |
| noise30_bit | .291752 | .265119 | Pass |
| noise30_xor2 | .282122 | .259808 | Pass |
| appears_bit | .269730 | .245181 | Pass |
| appears_xor2 | .267893 | .244538 | Pass |
| disappears_bit | .279049 | .258634 | Pass |
| disappears_xor2 | .282168 | .268573 | Pass |
| reverses_bit | .254828 | .219629 | Pass |
| reverses_xor2 | .263431 | .237050 | Inconclusive |

The last gain is .026382, with descriptive mean +/- 3.5 SE interval
[-.002743, .055506]. It fails the positive-lower-bound requirement. These
fixed-sample screens are not confidence sequences or research-wide coverage.
All first-cohort changed-case gain screens passed. The nonharm allowance is
.01, so passing it does not establish zero stationary harm.

Full inner learning is not uniformly better than outer-only: second-cohort
disappears_bit is .258634 versus .258422. Six second-cohort changed cases
remain above the .25 Brier of a constant .5 forecast. Relative improvement
does not establish adequate absolute prediction quality.

Foreground acquisition costs also rose: for example noise30_xor2 increased
from 4.392700 to 4.596558 coordinates/frame. Common audit and monitor costs
are recorded separately in the raw artifact. This is not an equal-total-cost
comparison against random or uncertainty sampling and does not validate goal 7.

## Reproducibility

- Contract: `mmm-arrival-breadth-v1-contract.md`.
- Data: `mmm-arrival-breadth-v1.jsonl`.
- Raw SHA-256: `9d15da0eaf451c7776caa9b8ed7316e87db187fbf102085bebf233935ca87089`.
- Summary: `mmm-arrival-breadth-v1-summary.json`.
- Evaluator: `research/arrival-breadth-summary.mjs`.
- 867 captured source hashes agree with current sources at verification.
- Independent recomputation checks all score/cost sums, immediate forecast
  equality, shared paired input tapes, and arrived/nonmissing audit fit origins.
- Summary regenerated on 2026-09-22 and compared byte-for-byte successfully.
- Contract test compares original-arm forecasts and metrics against the older
  driver for trajectory zero of every case and schedule, not every trajectory.

The prior collection checkpoint records a 17.430-second race contract test,
passing package vet, and 114.646-second package collection. This reporting pass
reran the independent evaluator, not the full Go collection. No new serving
performance benchmark was performed; collection time is not serving latency.

## Next Decision

Test the frozen method on the older majority/parity switching families where
full stale-role carry previously failed. This breadth study does not erase
those failures. Do not enlarge only the inconclusive cell until it passes.
Fresh confirmation, actual agent utility, equal-cost observation comparisons,
and integrated latency/freshness evidence remain outstanding. No production,
whitepaper, or remote repository changes were made.
