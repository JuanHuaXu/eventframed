# Credit monitoring in the learning loop

## Outcome: Quality Rescue Fails

0/4 frozen reverse-recovery gain screens pass;16/16 full/post nonharm screens
pass their .01 allowance. Better split detection does not establish better
forecasting. Do not adopt the candidate as a predictive-quality rescue.

Full arrival-learning control versus the SAME arm with credit monitoring:

| Cohort / feedback | Original post Brier | Credit post Brier | Mean gain | Gain interval |
| --- | ---: | ---: | ---: | --- |
| 1 / immediate | .231382 | .231338 | .000044 | [-.000047, .000135] |
| 1 / delayed | .259423 | .259748 | -.000326 | [-.002725, .002074] |
| 2 / immediate | .217915 | .218989 | -.001075 | [-.003495, .001345] |
| 2 / delayed | .253277 | .252457 | .000819 | [-.002018, .003657] |

Gain means control minus candidate. Requirements were mean>=.005 and positive
mean-3.5SE lower endpoint. These are16-trajectory descriptive fixed-sample
intervals, not repeated-research confidence coverage. Both delayed reverse
means still exceed neutral Brier .25. Stable predictions are unchanged. Delayed
forward changes are zero in cohort1 and a tiny .000021 gain in cohort2.

## Actual Sequential Integration

The research driver runs each forecast before its same-tick outcome. It uses
the same frozen credit step/cache and gate, not injected shadow decisions.
All512 credit transitions and split clocks reproduce the shadow experiment.
Latent inputs/outcomes, audit nominations, missing/delayed availability, released
origins, and every fit-origin manifest match the original. New monitor evidence
changes actual split state, model-slot selection, subsequent acquisition, and
issued forecasts. This is not just an offline relabeling of prediction scores.

All four internal learner arms are retained in raw data; the primary comparison
is arm2 against original arm2, not against the weaker old conservative arm.
Disabled-policy parity matches all four representative task cases under both
schedules exactly. No thresholds, priors, fit schedules or selector constants
were changed.

## Cost Boundary

Every monitor/audit prefix respects the credit guarantee. Foreground observation
choices can still change after different splits. Including foreground cost,
three of256 schedules exceed the original total by2,2, and4 coordinates across
512 frames. The monitor-only invariant is not a whole-system cost guarantee.
Mean total cost is slightly lower, but neither that nor unchanged per-query
foreground caps removes these individual differences.

## Influence Diagnostic

`mmm-credit-learning-v1-influence.json` describes only reverse-case frames where
the candidate has split but the original has not. It verifies issued forecasts
against the actual fixed-share weighted expert sum. Mean effective weight on
the changed pooled/local slot is2.12% in delayed cohort1 and6.06% in cohort2;
mean absolute forecast movement is .01962 and .03817 respectively. The neutral
slot carries65.30% and36.09%. Changes are real but do not yield consistent gain.
This is a descriptive conditional interval, not a causal mediation decomposition.

The earlier `mmm-gate-mediation-v1-report.md` diagnosed low slot influence, and
`mmm-local-birth-v1-report.md` / `mmm-fixed-observer-birth-v1-report.md` already
rejected fixed10% fresh-prior rescues. Do not repeat those as new discoveries or
sweep larger priors. The inherited action clips a raw slot weight before
renormalization; that operation is not a persistent strict .1 normalized cap.
Its semantics were preserved in this frozen comparison.

## Reproducibility And Performance

- Contract: `mmm-credit-learning-v1-contract.md`.
- Raw: `mmm-credit-learning-v1.jsonl`, SHA-256
  `4b148306f6fc1fa0f4e8ae5f1db3b878e3f1fd20b628f305313387a91f9af655`.
- Summary: `mmm-credit-learning-v1-summary.json`.
-128 trajectories, each with two schedules:256 schedule-runs,512 frames each.
- All128 records /256 runs replay exactly. The replay source snapshot includes
  a subsequently added benchmark, so this is record equality, not whole-file
  byte equality across differing metadata.
- All878 captured source hashes checked. Independent summary reconstructs all
  four arms' metrics, as-of training, costs, ledger prefixes and shadow clocks;
  regenerates byte-identically.
- Disabled-policy race test7.454s; package vet passes. Collection39.05s,
  full replay39.09s. No full-collection race claim is made.

Apple M4, whole512-frame/four-arm fixture, three one-iteration measurements:

| Schedule / policy | Times (ms) | Allocated MB |
| --- | --- | --- |
| Immediate / original | 179.716 /154.471 /157.276 | 71.86 |
| Immediate / credit | 154.091 /152.081 /156.306 | 71.90 |
| Delayed / original | 119.101 /121.932 /121.440 | 54.71 |
| Delayed / credit | 123.813 /119.452 /118.658 | 54.75 |

No meaningful speed superiority is established. This includes synthetic fitting
and all four learner arms, excludes initial base fitting and external service
I/O, and is NOT per-frame serving latency. The result does not meet goal6.

## Next Distinct Diagnostic

The credited monitor sometimes acquires all nine coordinates before forecasting,
yet `subsetState.predict` evaluates experts only on its own requested view. The
cached Reader correctly returns only requested coordinates; do not silently
break that contract to expose more. Investigate explicit reuse of already-paid
available evidence at the prediction boundary, with requested/available/consumed
masks and incremental costs distinguished.

First determine on these consumed trajectories whether the existing models
actually improve with those available coordinates, without fitting on future
labels or substituting hindsight weights. If that fails, representation/learning
needs improvement; observation reuse alone is not a rescue. Any new policy must
declare how the old six-coordinate foreground cap relates to shared cached data
and account for total work rather than merely changing the cost label.

All seven goals remain open. No production, whitepaper, remote repository or
private-data changes were made.
