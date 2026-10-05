# Arrival-time routed learning v106

**Overall FAIL: 302/318 frozen gates pass.** Removing prefix waiting from the
selector produces a small recovery improvement over full carry, but not the
required .005 gain. The candidate also fails to beat the conservative journal
on the delayed change cases. This is a partial effect, not an adoption pass.

## Evidence and integrity

- [Frozen protocol](mmm-arrival-v106-protocol.md)
- [Step-level artifact](mmm-arrival-v106.json):179,203,890 bytes, mode0600.
- [Independent summary and all318 gates](mmm-arrival-v106-summary.json)
- [Evaluator](../../research/arrival-v106-summary.mjs)
- [Driver](../../internal/observationlearners/arrival_v106_test.go)

768 independent latent trajectories each run under immediate and combined
delay/missing feedback:1,536 paired schedule-runs,393,216 scored frames. All runs
replay exactly. Paired latent tapes, all scores, as-of fit origins, journal
settlement and arrival-time selector counts reconstruct independently. The23
source/protocol/evaluator hashes match, and the summary regenerates byte for
byte. All v104 controls match exactly in consumed compatibility tests; immediate
arrival and full carry are identical. Race checks, seed checks and vet pass.

Generation185.56 seconds; complete replay187.49 seconds. Artifact SHA-256:
`bc2d8289e2ed1910c7fdf0a425265453f2ea5e1b282b8687622624fb1fdf9d45`.
Distinct effective seeds are audited against v90-v104 allocations; v105 reused
v104. This is not an independent generator or a guarantee across research history.

## Confirmation quality

Late128 under jitter0..31/missing0.2. Expected Brier, lower is better:

| Case | Generic64 | Conservative | Full carry | Arrival selector |
| --- | ---: | ---: | ---: | ---: |
| Parity4 | .072410 | .069955 | .060787 | .060069 |
| Majority to parity | .254168 | .250863 | .252616 | .251484 |
| Parity to majority | .228414 | .227753 | .232110 | .230627 |

Arrival's expected accuracies for these rows are94.76%,66.70% and69.58%.
Stable-parity performance is retained, but the two change cases remain much
weaker. Majority-to-parity Brier still exceeds the constant-.5 score of .25.
That arithmetic reference is diagnostic, not an extra frozen gate.

Late arrival gain over full carry on the confirmation change cases:

| Case | Mean gain | Approximate paired interval | Frozen requirement |
| --- | ---: | --- | --- |
| Majority to parity | .001132 | [.000018, .002245] | mean>=.005, lower>0 |
| Parity to majority | .001483 | [.000396, .002571] | mean>=.005, lower>0 |

The direction is favorable, but both fail the practical gain floor. Design
gains are .000913 and .001484; the first design interval also crosses zero.
This does not support a larger-size replication merely to erase the failure.

Relative to conservative, confirmation arrival gains are -.000621 and -.002873
on these two cases. Their intervals cross zero, but neither demonstrates the
required improvement. Design intervals show adverse effects relative to
conservative. Full carry's v104 failure remains standing.

## Gate accounting

- All288 non-harm gates pass their upper-harm<=.01 allowance.
- 12/20 generic-comparator gains pass. All four parity3 cells miss the .005
  mean floor, and all four delayed switch gains fail.
- 2/6 gains over conservative pass: delayed late parity4 in each phase.
- 0/4 additional switch gains over full carry pass the .005 mean floor.

A passed .01 harm allowance is not zero harm. Intervals are the frozen paired
mean +/-3.5 sample standard errors over32 trajectories. They are approximate
fixed-sample screens, not confidence sequences or a correction for every past
research decision. Paired feedback schedules are not independent replicates.

## What changed and what did not

The selector applies a captured loss at label arrival, without waiting for a
missing origin prefix. The evidence gate retains its original ordered release
and version restrictions. Priors, per-label fixed share, model fitting, label
availability, acquisition rules and gate mathematics are unchanged.

The independent checker verifies that arrival-time selector counts follow
actual availability, while gate counts follow the resolved prefix. Final
settlement totals match full carry; timing differs. No stale loss enters a
new-version gate and no loss is applied twice to the selector. Full-frame
training audits remain separate from the six-coordinate foreground cap.

The experiment isolates an operational timing change, but its downstream
forecast effect includes both changed selector weights and changed observation
choices. It does not establish that age discounting or event-clock forgetting
would solve the remaining error.

## Cost

[Benchmark output](mmm-arrival-v106-benchmarks.txt), Apple M4 darwin/arm64,
suffix -10, three one-iteration measurements after replay:

| Complete four-arm256-frame learned fixture | Time | Allocated bytes | Allocations |
| --- | --- | --- | --- |
| Immediate | 120.179-146.440 ms | 26.478 MB | 4,469-4,470 |
| Delay/missing | 117.786-119.695 ms | 26.511-26.516 MB | 4,611-4,617 |

This includes fitting, publication and final flush, not just one forecast.
It excludes database retrieval, network service and production OpenClaw. Do
not call these request latencies or compare them directly to differently sized
earlier multi-arm fixtures. The prior paired component measurement found9-11%
extra bookkeeping for arrival selection, with unchanged allocations.
Benchmark source SHA-256:
`7deae5c093085585f556f60ca49d2a46186b60758bc3eade445cb90e1023dd29`.

## Next discriminating check

Before changing more selector constants, measure the recoverable error from
the existing fitted forecasts. A read-only check of the consumed v105 profiles
finds that, at the change clock128, all four pure forecasts have full-input
Brier>.25 in all32 confirmation trajectories of each change direction. This
is not proof that every convex mixture is bad: mixtures can reduce Brier even
when each component is individually worse than neutrality.

Therefore calculate a consumed-data convex-mixture oracle bound, both with and
without the neutral forecast, using verified constrained optimization. Separate
early lack of useful forecasts from later failure to select useful ones. The
oracle is diagnostic only and must never supply runtime weights. Then choose
the next separately frozen forgetting, discounting or neutralization ablation.
Do not treat pure-model averages as a mixture impossibility theorem.

All seven directions remain open. Broader v88/v80 requirements, independent
generators, real-agent evidence and persistence tails are not replaced by this
screen. No production changes, whitepaper edits, commits or pushes were made.
