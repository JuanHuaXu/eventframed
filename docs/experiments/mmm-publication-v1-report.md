# Publication holdout does not rescue recovery

**FAIL.** Warm calibration improves the heldout-reset control, but fails against
the unchanged learner. All16 changed-case protection comparisons with the
unchanged learner fail; immediate regressions are substantial. Do not promote
this candidate or present its weaker-control wins as a successful rescue.

## Evidence

- [Frozen contract](mmm-publication-v1-contract.md)
- [Raw forecasts, fit partitions and source snapshots](mmm-publication-v1.jsonl)
- [Independent summary](mmm-publication-v1-summary.json)
- [Byte-identical summary replay](mmm-publication-v1-summary-replay.json)
- [Independent evaluator](../../research/publication-summary.mjs)
- [Training-age diagnosis](mmm-publication-v1-age.json)

192 independent trajectories under two paired feedback schedules, three paired
arms per schedule. Do not count schedules or arms as independent trajectories.
All512 issued forecasts per arm are scored, including labels hidden from the
learner. All840 captured source snapshots match. The evaluator reconstructs
scores, generation settlement, arrival eligibility, fit partitions, literal
calibration weights and their appearance in the next served forecast.

Complete Go replay reproduces every384 schedule-run, including all three arms.
Collection61.26s; replay62.01s. The replay/benchmark file was added after the
collection snapshot and is not part of those840 hashes. Compatibility race tests
passed in4.222s before collection; package vet passed.

Raw SHA-256: `566710fe18c30d2fa905b4c5a9203874bf25d965b73da846c15cff2efcddc596`.

## Gate accounting

- Warm versus unchanged:8/24 non-harm screens,0/8 primary delayed gain screens.
- Warm versus heldout-cold:24/24 non-harm screens,8/8 primary delayed gains.
- Combined:32/48 non-harm and8/16 delayed gain comparisons pass. Overall FAIL.

Each non-harm screen covers full and post-change scores with lower gain>=-.01.
Gain requires post-change mean>=.005 and a positive lower bound. Intervals use
paired mean +/-3.5 standard errors over16 trajectories. They are approximate
fixed-sample screens, not confidence sequences or guarantees over search history.

## Second cohort

Post-change mean Brier (lower is better):

| Case | Schedule | Unchanged | Heldout cold | Heldout warm |
| --- | --- | ---: | ---: | ---: |
| Copied bit | Immediate | .194749 | .239361 | .232427 |
| Copied XOR2 | Immediate | .212278 | .249193 | .238967 |
| Noisy-copy bit | Immediate | .201917 | .241968 | .233672 |
| Reversing XOR2 | Immediate | .180795 | .226834 | .217563 |
| Copied bit | Delayed | .279590 | .319943 | .281767 |
| Copied XOR2 | Delayed | .279574 | .319049 | .281333 |
| Noisy-copy bit | Delayed | .280556 | .329839 | .286522 |
| Reversing XOR2 | Delayed | .273903 | .319326 | .274110 |

For example, delayed copied-XOR2 warm gain over cold is.037716,
interval[.025744,.049688]. Against unchanged, gain is-.001759 with
interval[-.013568,.010049]. The first result is real but not an end-to-end win.

All second-cohort immediate changed-case warm gains versus unchanged are
negative, with intervals excluding zero; harm means range.02669-.03768.
Delayed changed-case gain intervals cross zero, but do not meet either the gain
requirements or the allowed lower-bound protection. Stable/null screens pass the
.01 allowance, which must not be described as exactly unchanged performance.

## What the failure does and does not establish

Sixteen withheld audited labels are not sixteen event frames. Across all
post-change publication records, excluding final-flush fits, the newest available
training origin is64.25 frames older under holdout with immediate labels and
85.25 frames older with delayed/missing labels. In the delayed run, mean age of
the newest training origin grows from10.21 to95.47 frames. The
[diagnostic script](../../research/publication-age-diagnostic.mjs) reconstructs
these descriptive counts from fit partitions; correlated fits are not treated
as independent evidence or used for confidence intervals.

Cold and warm also reset selector weights at every publication. Therefore this
experiment does not identify how much harm comes from withholding training
evidence versus erasing accumulated selection evidence. Warm/cold forecasts can
choose different observations, so their difference includes observation effects.
The validation labels never enter the calibrated model fits; no optimistic
in-sample refit was substituted to rescue the result.

Next, diagnose data withholding and selector reset separately on consumed data
before proposing a new policy. Preserve original endpoints and forward scores.
Do not shrink the holdout or tune the prior on these outcomes and label that
confirmation. Any later rescue still needs fresh cohorts and both controls.

## Cost

Apple M4 darwin/arm64; three complete three-arm512-frame fixtures:

| Schedule | Time | Allocated bytes | Allocations |
| --- | --- | --- | --- |
| Immediate | 216.937 /189.485 /189.740ms | about168.126MB | 54657 /54656 /54661 |
| Delayed | 141.810 /142.145 /141.733ms | about126.020MB | 51673 /51678 /51680 |

Command: `go test ./internal/observationgate -run '^$' -bench '^BenchmarkPublicationFixture$' -benchtime=1x -count=3`.
All measurements are retained. These include all arm fits and calibration, not
one request; they exclude network/storage service. Cold computes but discards
calibration weights to match work. More arms and separately fitted bundles mean
these costs cannot be directly called overhead relative to the earlier two-arm
forest fixture. No serving-latency or freshness success follows.

All seven goals remain open. No production, whitepaper, commit or push changes.
