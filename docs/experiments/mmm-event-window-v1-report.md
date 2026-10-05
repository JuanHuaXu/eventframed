# Event-time short window: model benefit, substitution FAIL

## Frozen test

Implemented the [event-window contract](mmm-event-window-v1-contract.md) as an
isolated test-only model intervention. The last64 event origins replace the
last64 audit labels ONLY in a shadow short count/subset pair. Original audit
eligibility, fit clocks, issued masks, inner/outer weights, long/base forecasts
and Anti-Pigeon clocks remain fixed. No selector learns from shadow outputs.

All128 trajectories, both schedules and both views are retained. Before scoring,
the experiment reconstructs the original short and complete forecast on every
frame, to tolerance1e-12. Fitting follows the current clock's forecast. Empty
event-window support produces neutral advice; missing/future/nonaudit support
is rejected. No true change clock enters the fitter.

**FAIL:0/4 reverse-recovery gains and12/16 nonharm screens on EACH view.**
All four immediate changing cells fail nonharm. The16-cell screens use full and
post loss; the gain screens require mean>=.005 and mean-3.5SE>0. Trajectories,
not frames, are the16 replicates. Consumed exploration, not fresh confirmation.

## Requested-view results

| Cohort / schedule / change | Original post Brier | Event-window post Brier |
| --- | ---: | ---: |
| 1 / immediate / majority to parity | .243420 | .258183 |
| 1 / immediate / parity to majority | .231338 | .240588 |
| 2 / immediate / majority to parity | .240403 | .255633 |
| 2 / immediate / parity to majority | .218989 | .237653 |
| 1 / delayed / majority to parity | .273706 | .273941 |
| 1 / delayed / parity to majority | .259748 | .258994 |
| 2 / delayed / majority to parity | .270235 | .269267 |
| 2 / delayed / parity to majority | .252457 | .255407 |

Delayed reverse gains are+.000754 and-.002950, with exploratory intervals
[-.003376,.004885] and[-.009678,.003778]. Both miss the frozen gain target.
The available-view endpoint also fails; the full table is in the summary.

## What we learned

The raw subset model DOES benefit on reverse shifts. On requested inputs,
delayed reverse post Brier improves .26537->.23584 in cohort1 and
.24896->.23432 in cohort2. On available inputs it improves
.27829->.23107 and .27373->.22817. These means beat neutral.25, so it is not
merely returning neutral on every example. They are not separate confirmed
statistical quality claims or calibrated posterior guarantees.

The support cost is severe on stable data: cohort2 requested-view stable
majority subset loss worsens .06026->.20934; parity worsens .19536->.26190.
The complete stable forecast largely survives because other original components
still dominate. That does not make the shortened model a safe replacement.

Better average constituent loss does not imply a better complete mixture.
Original weights were learned from different issued advice and may emphasize
the wrong examples or models after substitution. Correlations between component
errors also matter. This experiment does not identify which mechanism dominates.
It falsifies the simple fixed-weight substitution rescue, NOT the possibility
of keeping both windows with honestly learned issued-advice weights.

Next inspect previous short-window-bank failures before testing competing
sample-count/event-time windows with separate, origin-bound loss records. Keep
the long-support model; do not sweep the event horizon, force more short-model
weight, or claim a gain against a weakened control. Start with model-level
competition, then test changed acquisition/feedback in a closed loop if justified.

## Verification and cost

- Race contract test PASS (1.296s package), plus vet. Tests cover window endpoint
  inclusion, empty support, invalid origin order, future/missing/nonaudit labels,
  and valid probabilities across all512 masks.
- Full collection20.87s; independent full native replay18.69s. Entire raw files
  are byte-identical. Full collection was NOT run under the race detector.
- Independent JavaScript checks1,572,864 scalar model-mixture/reference values,
  max difference3.33e-16, plus every fit's retained origin sets and publication
  ordering. It verifies304 captured source/contract hashes. It reconstructs
  mixture arithmetic, not an independent implementation of the subset fitter.
- Summary replay is byte-identical. Input artifacts are exact-hash bound.
- Apple M4 count-plus-subset fit benchmark:10 labels5.535-5.547ms versus64 labels
  6.374-6.447ms; both allocate about2.75MB and5 allocations per fit. Keeping both
  pairs adds a fit and retained model tables; no memory-saving claim.
- Count/subset forecast lookup8.29-8.68ns,0allocations. This excludes aggregation,
  selection, database I/O, acquisition, publication and serving contention.

No production behavior, whitepaper, remote, preexisting tracked edit or claim
register changed. All seven research goals remain open.

## Artifacts

Raw `mmm-event-window-v1.jsonl` and `mmm-event-window-v1-replay.jsonl` SHA256:
`5d5c0d8b4fc0447f738311019f4475387cf51f81969733578942335ab189b332`.
Summary `mmm-event-window-v1-summary.json`; evaluator
`research/event-window-summary.mjs` and fitter source are captured in raw header.

Post-collection benchmark source SHA256:
`57956b6fdde105bc858235231a1e17202d39bc35eaf52e9509d3cd54c254cc56`.
It is separate from the original collection source manifest.
