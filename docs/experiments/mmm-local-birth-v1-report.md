# Fresh local prior: failed factorial pilot

## Decision

Do not promote. Resetting the newly activated local expert to its existing10%
prior does not deliver the required forecast gain. Mean effects change sign for
member shifts and are adverse for recurring changes in both cohorts. The
diagnosis of inherited low weight remains correct, but giving the slot more
influence is not itself a successful rescue.

[Frozen protocol](mmm-local-birth-v1-contract.md),
[raw records](mmm-local-birth-v1.jsonl),
[summary](mmm-local-birth-v1-summary.json),
[replay](mmm-local-birth-v1-summary-replay.json).
Raw SHA-256:
`1c9dd2478e8b4c075f079b0082a7f7cff8b8ddc3b82bdb5ec390ea0b939bf926`.
All809 captured source hashes match embedded/current files; analysis replay is
byte-identical. No full trajectory replay is claimed for this new artifact.

## Action and accounting

At the first prediction after authorized splitting with a local model available,
the test-only action sets the local selector slot to.1 and proportionally rescales
the other slots to sum.9. It runs once, preserves their relative learned weights,
and does not score the local expert retrospectively on the revealing outcome.
Failed prediction restores the prior mixture weights; invalid prediction order
is rejected before activation. The normal implementation remains unchanged.

Four paired subset arms cross old/new gate with inherited/birth action. Five
original count-workflow controls remain. The160 trajectories span five scenarios,
two fresh cohorts and16 independent4096-label fits per cell, each with512 frames.
Models/audit labels are shared immutably across paired arms, but forecast and
weight histories are separate. This is a small exploratory pilot, not broad
false-revocation, delayed-feedback or dependence validation.

## Primary contrast: birth versus inherited under mixture gate

Gain is control minus candidate post-Brier; positive is better. Intervals are
paired trajectory mean +/-3.5SE, not confidence sequences or simultaneous bounds.

| Cohort/scenario | Inherited Brier | Birth Brier | Mean gain | Interval |
| --- | --- | --- | --- | --- |
| Design/member | .203604 | .205453 | -.001849 | [-.004523,.000825] |
| Second/member | .203495 | .202657 | .000839 | [-.001345,.003022] |
| Design/recurring | .176466 | .178524 | -.002058 | [-.006141,.002025] |
| Second/recurring | .168033 | .168815 | -.000782 | [-.002334,.000770] |

No positive-gain screen passes the unchanged conjunction of mean>=.005 and
positive lower bound. All primary full/post .01 non-harm screens pass, which
does not prove benefit or safety for arbitrary streams. Stable, common-shift
and null predictions are identical because no authorized split activates birth
in those cells. Their zero differences do not establish rare-event coverage.

Within the birth action, faster gating also fails the .005 gain screen. Member
gate effects are negative in both cohorts; recurring changes sign. Thus the
factorial comparison does not hide a useful gate effect behind the action mean.
Raw summaries retain every arm and both full/post contrasts.

## Work and verification

Second-cohort member observation cost increases4.60327 ->4.66577 coordinates per
frame; recurring4.86462 ->4.89368. Each call still respects the six-coordinate
cap. Fixed fitting/audit budgets do not imply equal realized observation work.
The action can also change which expert guides acquisition, so the prior
fixed-mask diagnostic cannot predict its full closed-loop effect.

Focused five-scenario control/disabled/action tests pass under race detection
(4.800s), including normalization and relative-weight invariants. Vet passes.
Collection completes17.62s (17.778s package), including all arms and fits.

Three500ms fixed-input arithmetic microbenchmarks on Apple M4, darwin/arm64,
report6.123,5.970 and6.011ns/op,0bytes and0allocations. Command:

```sh
go test ./internal/observationgate -run '^$' -bench '^BenchmarkLocalBirthWeights$' -benchtime=500ms -count=3 -benchmem
```

This measures only the four-weight transformation, not the activation branch,
observation, fitting, persistence or serving p99. Low arithmetic cost does not
rescue the failed quality result or erase additional acquisition cost.

## Next boundary

Preserve the failure; do not sweep larger birth weights on these outcomes. A
distinct diagnostic can hold the observation policy fixed while changing the
forecast mixture, separating acquisition changes from local-model influence.
It must use the same permitted observed mask, never additional hidden features,
and retain actual issued forecasts before feedback. First inspect existing
observer/forecast decoupling experiments to avoid repeating a failed variant.

No specialist-regret, posterior calibration, runtime promotion or goal-completion
claim follows. All seven goals remain OPEN. Production, whitepaper and remotes
remain untouched.
