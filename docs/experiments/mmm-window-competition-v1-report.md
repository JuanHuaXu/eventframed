# Window competition: partial recovery, not complete

## Verdict

**2/4 reverse-recovery gains;16/16 nonharm screens, on each view.** Immediate
reverse cases pass. Delayed cases do not. Allseven research goals remain open.
The existing .005 minimum gain is unchanged; near misses are still failures.

The [frozen contract](mmm-window-competition-v1-contract.md) retains original
short advice and adds event-time count/subset forecasts through a new inner
ForecastMix. A separate outer ForecastMix learns its own actually issued advice.
Updates occur at label arrival, with original external split timing and pending
origin rejection. Acquisition and fitted component tapes remain fixed. This is
not a closed-loop replacement for the operational observer.

Primary requested-view reverse results (post-window Brier, lower is better):

| Cohort / schedule | Original | Competition | Gain | Exploratory interval | Pass |
| --- | ---: | ---: | ---: | --- | --- |
| 1 / immediate | .231338 | .224145 | .007193 | [.002083,.012303] | Yes |
| 2 / immediate | .218989 | .208538 | .010451 | [.002922,.017981] | Yes |
| 1 / delayed | .259748 | .254950 | .004798 | [.000479,.009118] | No |
| 2 / delayed | .252457 | .248096 | .004362 | [-.000474,.009197] | No |

Intervals are paired mean+-3.5SE over16 trajectories, not anytime certificates
or independent-frame confidence intervals. All data are consumed exploration.
Delayed cohort1 remains worse than neutral.25; cohort2 mean just beats it.

Immediate reverse accuracy improves60.52%->65.23% and62.72%->67.70%.
Delayed reverse improves53.13%->57.20% and55.03%->57.28%. Stable cases are
largely unchanged, with very small probability differences. Immediate forward
cases incur mean harms.001683 and.001990, within the declared nonharm tolerance,
not literally zero harm. Preserve those costs when interpreting the rescue.

## What changed relative to replacement

The event-time models receive roughly57-63% inner mass during requested-view
reverse cases, but only~.3-.5% during stable majority cases. Evidence-based
competition can protect the older well-supported model while using fresher
advice after changes. Both inner and outer prediction weights are updated;
this experiment does not separately identify their causal contributions.

Previous v97's shorter-window bank also failed and remains a negative result.
This sparse-audit, delayed-arrival, nested count/event-time competition is a
different bounded experiment, not a retuned claim that v97 passed.

Available-view reverse gains are.014633/.023276 immediate and.012517/.016906
delayed. Delayed intervals cross zero, so both fail. Crucially, that comparator
is the original AVAILABLE law using NARROW-trained weights. It is not the actual
original serving forecast. For example, cohort2 delayed available candidate
.252347 is almost the same as original requested control.252457, not a.0169
gain against it. Requested-view results above remain primary.

Next: preserve this candidate without changing its thresholds, then establish
whether the benefit survives acquisition/feedback coupling. Before implementation,
inspect the conditional-law observer interfaces: choosing a view from one model
while scoring another needs an explicit contract. Delayed recovery remains open;
do not select fresh seeds until a consumed run happens to pass or tune prior/share
against these cells. A fixed-view component benefit is not goal2 or goal7 success.

## Verification and audit correction

The first collection stopped at frame0: its checker compared raw zero-state
weights with the tape's serialized default prior. `subset_state.go` explicitly
serializes [.7,.1,.1,.1] when the internal weights are zero. Corrected only that
representation comparison in Go and the independent evaluator, and changed the
unit fixture to exercise the real serialized form. No score was produced or
used for tuning. The failed header `mmm-window-competition-v1.jsonl` is retained;
it is NOT a complete experimental artifact. Use the `-verified` file below.

- Corrected race journal test PASS (1.250s package), vet PASS. Tests isolate
  future-label effects, missing labels and invalid arrival clocks.
- All256 schedules reconstruct original outer weights/laws at every forecast,
  including before/after external split transitions. Existing pre-normalization
  revoke cap is preserved; no permanent .1 bound is claimed.
- Collection1.79s; full race collection28.55s. Entire verified/race raw artifacts
  are byte-identical. This includes512 forecasts per view per schedule.
- Independent JavaScript checks5,505,024 scalar advice/weight/forecast values,
  maximum difference1.11e-15. Seven focused source/contract hashes verify.
- Summary replay is byte-identical; three parent tapes are bound by exact hashes.

## Cost and artifacts

One view's two-level forecast/update kernel on Apple M4:66.56/66.65/66.00ns,
0B and0allocations. This excludes fitting, acquisition, journal storage, I/O,
publication and contention. The extra fitted pair still costs about5.5ms and
2.75MB allocation per fit as measured in the event-window experiment. Cached
advice replay does not remove that cost from a real implementation.

Authoritative raw: `mmm-window-competition-v1-verified.jsonl`; full race replay:
`mmm-window-competition-v1-race.jsonl`. Both SHA256:
`ecb3cf7f99d7f1460499c4d4ba9e6473530c47b8d824a7ee61ac386305ad640b`.
Summary: `mmm-window-competition-v1-summary.json`; captured evaluator:
`research/window-competition-summary.mjs`.
Post-collection benchmark source SHA256:
`342947b1464a148bcca4cdd2ce8a05b7649b584df46a24e511627961802b08ee`.

Production, whitepaper, remotes and existing tracked changes remain untouched.
No adoption, fresh confirmation, full-service latency or completed-goal claim.
