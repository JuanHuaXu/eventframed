# Arrival-learning switch transfer

## Outcome

Not rescued: 2/4 delayed switch gain screens pass, and 16/16 full/post nonharm
screens pass their .01 allowance. Both majority-to-parity cells pass; both
parity-to-majority cells have positive means but intervals crossing zero.
All seven research goals remain open. No learner constants were tuned.

| Cohort | Direction | Original post Brier | Full arrival | Gain interval |
| --- | --- | ---: | ---: | --- |
| 1 | Majority to parity | .288759 | .273706 | [.003467, .026639] |
| 1 | Parity to majority | .266072 | .259423 | [-.000645, .013943] |
| 2 | Majority to parity | .289677 | .270256 | [.006206, .032635] |
| 2 | Parity to majority | .258939 | .253277 | [-.000606, .011930] |

All four switched-case full-arm means exceed constant-.5 Brier .25. Passing
relative gain is not sufficient absolute quality. Intervals are descriptive
mean +/-3.5 SE over16 trajectories, not confidence sequences or coverage over
all historical research decisions. Stable cases remain controls, not gain targets.

In cohort2 delayed stable_majority3, original/full post Brier is .051146/.051261;
stable_parity4 is .182847/.183083. Small harm exists despite passing nonharm.
For majority-to-parity, full is slightly worse than outer-only (.270256 versus
.270121); parity-to-majority full improves on outer-only (.253277 versus .254637).
This does not support a universal inner-update benefit.

## Design Boundary

128 exploratory trajectories, each with immediate and delayed/missing schedules:
256 schedule-runs, four arms,512 frames each. This transfers the old accessible
majority/parity rule families to the current external-gate architecture; it is
NOT an exact v104/v106 replication. Unlike those older studies, the current
architecture has4096 pre-regime base samples,25% audit nomination, a member-only
shift at256, and an unchanged paired reference. The rule masks are not supplied
to the learner. Both alternating accessible-mask pools are tested. Prior
v104/v106 failures remain standing, and neither cohort is called confirmation.

The test-only callback supplies latent outcomes after predictions. Default
behavior is unchanged; an explicit old-label callback reproduces the default
driver exactly under both schedules. Every subsequent fit still uses only
arrived, nonmissing audit labels. Role-credit handling, priors, caps, and gate
authority remain unchanged. No serving implementation was edited.

Foreground cost in cohort2 delayed majority-to-parity is4.515015 versus4.528442
coordinates/frame (original/full); parity-to-majority is5.996582 versus5.995239.
Audit and monitor costs remain recorded separately. This is not an equal-cost
comparison against random/uncertainty observation and does not validate goal7.

## Post-Hoc Diagnostic

`mmm-arrival-switch-transfer-v1-oracle.json` uses the synthetic latent truth to
compute the per-frame best convex mixture of the issued full-arm forecasts.
It knows the answer and can change weights each frame: an unattainable lower
bound, NOT a trained policy, calibrated posterior, or measured rescue. The
neutral forecast is already one of the four experts.

For cohort2 delayed (expected rather than realized Brier):

| Direction / frames | Actual | Truth-informed oracle | Neutral optimal fraction |
| --- | ---: | ---: | ---: |
| Majority to parity /256-383 | .291008 | .150808 | 47.51% |
| Majority to parity /384-511 | .249683 | .152244 | 39.60% |
| Parity to majority /256-383 | .270521 | .174127 | 27.88% |
| Parity to majority /384-511 | .238028 | .154149 | 18.51% |

There is selection headroom on issued views, but substantial times when neutral
is best among them. The calculation does not identify a learnable selector or
counterfactual observation policy. The external gate split16/16 majority-to-parity
trajectories but only1/16 parity-to-majority trajectories in this cohort.
This asymmetry motivates inspecting the correctness-based monitor's sensitivity
to conditional changes; it is not proof that loosening the gate would help.

## Verification And Cost

- Raw: `mmm-arrival-switch-transfer-v1.jsonl`.
- SHA-256: `c21f8ba8014956beecdf4866fff87efab393ff7b5f0dbe233747fd03840c9775`.
- Independent summary: `mmm-arrival-switch-transfer-v1-summary.json`.
- All868 captured source hashes checked; complete256-run Go replay byte-identical.
- Summary regenerated and compared byte-for-byte. Reconstructs scores, costs,
  paired tapes, immediate-arm equality, and as-of fit availability.
- Race contract test passed in2.662s; package vet passed.
- Collection38.65s; full replay38.99s, excluding package startup.
- Oracle script is a postcollection diagnostic, not in the collection snapshot.

Unchanged whole512-frame/four-arm inner fixture benchmark, Apple M4:
immediate176.912/154.095/153.907ms,71.55MB allocated;
delayed133.174/132.092/133.980ms,63.12MB allocated.
These are total fixture times, excluding initial base fitting, not serving
latencies or benchmarks of the new switch tasks. They resemble the earlier
fixture measurements but do not establish production nonregression.

The generic test-driver extraction changes a file captured in older experiments.
Those snapshots remain authoritative for their runs; older live-source hash
checks should not be expected to match this later revision.

## Next Step

Audit split sensitivity on these consumed trajectories: distinguish lack of
observable conditional evidence from loss of information in scalar correctness
monitoring. Freeze any proposed replacement and its null/error-control tests
before another rescue run. Do not tune thresholds or extend only failed cells.
No production, whitepaper, remote repository, or private-data changes.
