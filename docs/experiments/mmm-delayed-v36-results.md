# Delayed-evidence identity V36: results

2026-10-03. **Frozen technical component PASS in BOTH splits. Model-quality
rescue NOT established.** All seven whole goals remain OPEN. Production,
whitepaper, prior protocols and raw cohorts are unchanged.

## Scope and Evidence

The new single-owner research adapter accepts several pending trials/member,
retains private original forecasts, resolves legitimate out-of-order labels
once, and distinguishes cancellation from negative evidence. Old owner/epoch,
replay, backward time and cap violations are rejected atomically. Explicit
epoch rotation resets the model; no changepoint detector is implemented here.
Tickets bind trial identities, not authentic or independent evidence sources.

Fresh design/confirmation bases 2026103603/2026103604 give 384 worlds,
2,304 schedule arms, 8,448 distinct snapshots and 921,600 distinct trial
observations. Six learner copies of a trial are NOT six independent labels.
Two geometries, six regimes and 16 independent worlds/cell cover aligned,
two-rate, curved, shared-rate, noisy-label and mid-stream reversal worlds.
Schedules cover immediate, fixed, random, burst, fully deferred reverse order,
and outcome-dependent arrivals. The last is explicitly a negative control
outside the ignorable-delay assumption, not a valid full-stream posterior.

Both full audits reconstruct RNG populations, original issued forecasts,
every receipt and all metrics. Separate batch Beta integrals/direct atom
products verify each snapshot against its actual arrived evidence set. Every
fully drained schedule agrees within 3e-10 on the same full-evidence forecast.
This is stationary-model algebra, not proof that a shifted/noisy law is right.
The independent math is within Go; RNG replay uses the frozen generator.

Thirteen source hashes are sealed in each manifest. Postcollection auditor
and benchmark hashes are recorded separately. Seven corruptions (issued law,
receipt, outcome, future law, risk, arrival time, cost) plus source mutation
are rejected. Future flips cover all six schedules, including an additional
outcome-coupled prefix test. Input ownership, cap, cancellation, epoch,
normalization failure and original forecast binding tests pass. Four research
modules pass race tests and vet.

Artifacts: [protocol](mmm-delayed-v36-protocol.md),
[preflight](mmm-delayed-v36-preflight.md),
[design audit](mmm-delayed-v36-design-audit.json),
[confirmation audit](mmm-delayed-v36-confirmation-audit.json),
[benchmarks](mmm-delayed-v36-benchmarks.txt). Raw JSONL files accompany them.

SHA256:

- design: `887933172a91fe6ca67193b94ad7857e07102ed14e39dccc2cd07d0e01e5fca3`
- confirmation: `e572adc4350cda67f93281f7cbb20d78010867c08ba7b62190fa551776e8b4da`

## Actual Prediction Quality

Confirmation wide-geometry means at tick2399, BEFORE draining late labels:

| Regime/schedule | Original issued expected Brier | Current future Brier | Top10 usefulness |
| --- | ---: | ---: | ---: |
| Independent / immediate | .180656 | .160726 | .800000 |
| Independent / fixed150 | .190534 | .160665 | .800000 |
| Independent / uniform299 | .186339 | .160795 | .800000 |
| Independent / outcome_coupled | .263424 | .162814 | .800000 |
| Mid-shift / immediate | .250109 | .249999 | .477500 |
| Mid-shift / fixed150 | .273447 | .250015 | .395000 |
| Mid-shift / uniform299 | .268451 | .250019 | .391250 |
| Mid-shift / outcome_coupled | .349903 | .250855 | .278750 |
| Noisy labels / immediate | .193429 | .168083 | .800000 |

Original issued expected Brier scores forecasts BEFORE their feedback arrives
against the true contemporaneous usefulness rates. Snapshot Brier scores the
current forecast for future trials. These are different, noninterchangeable
objects; the full table records both and all other schedules/checkpoints.

Relative issued-Brier harm versus immediate in independent worlds:

- Fixed150: +.009877, interval [.009334,.010420].
- Uniform299: +.005683, interval [.004306,.007059].
- Outcome-coupled: +.082767, interval [.079880,.085654].

These are paired mean +/-3.5SE intervals over16 worlds, not time-uniform
coverage or a data-selected adoption gate. Positive values mean harm.
The design cohort independently exhibits the same direction.

Outcome-dependent arrival can badly bias the learning path while its final
snapshot looks good. Observing positives promptly and negatives late is not
ordinary unselected evidence. A selection/delay joint model or a valid envelope
is required; original forecast scoring prevents a retrospective endpoint from
hiding the harm. No delivered quality improvement is claimed for this control.

After the mid-stream reversal, full-history pooling averages contradictory
old/new evidence rather than recovering the new regime: future Brier stays
near .25 versus the true-rate floor .16. Noise also exposes a gap: independent
10% label flips yield packed signed bias -.052059 in the immediate arm.
That is not resolved by identity correctness. Recovering current regimes and
latent usefulness needs new, prospectively tested window/noise mechanisms.

## Cost

Apple M4, Go1.27.1,150 members: Issue244.9-263.2ns, Resolve10.037-10.104us,
both zero allocations. These microbenchmarks include restoring benchmark
state; they are not network, acquisition or serving timings. Constructor
77.416-80.044us,740276-740277 bytes/10 allocations. Epoch reset
66.654-67.829us,500997 bytes/six allocations. Storage is O(N*378+N*64),
not constant in frontier size.

Maximum accounted learner phase sum41.729ms design/39.856ms confirmation
passes the declared50ms component cap. Schedule building is measured separately;
scoring is evaluator work. There is no loaded service, persistence, queue age,
source authentication or production-clock claim.

## Auditor Repair and Next Direction

The first design audit failed its own fixed expected snapshot count:
4224 observed versus4416 expected. It had counted coincident2399/final
snapshots once for immediate but twice for burst. Both schedules legitimately
finish at2399. The checker now derives the checkpoint union from independently
audited arrival times; both full audits were rerun successfully. No candidate,
protocol source, outcome, quality gate or recorded forecast was changed.
The protocol's coincident-final wording applies to burst as well as immediate.

Next test bounded rolling/adaptive evidence windows and explicit delay/noise
selection contracts on fresh cohorts. Keep full-history and fixed-window
controls, original forecasts and delayed receipts. A correct adapter is not
Goal1 robustness, Goal2 recovery, AP split authority, Goal4 shifted efficiency,
Goal5 agent improvement, Goal6 loaded freshness or Goal7 total-cost superiority.
