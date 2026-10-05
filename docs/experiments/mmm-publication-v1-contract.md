# Frozen publication-holdout quality test

192 independent latent trajectories: six forest-delay cases,16 fresh4096-label
bases per cell, cohorts2027010351/2027010451. Each runs immediate and delayed
feedback. Three arms: unchanged uniform-input learner; heldout fit with cold
default selector; heldout fit with calibrated outer selector. All use uniform
input marginalization, same raw evidence, audit schedule and external AP gate.

At a32+16k arrived-audit fit threshold, reserve newest16 origins for validation
and fit on older origins only. Preserve the calibrated fitted objects, no refit.
Cold and warm both reset inner to default .7/.1/.1/.1; cold outer also defaults,
warm outer uses declared Bernoulli holdout weights. Calibration uses each arm's
historically issued observed masks. No new acquisition, but charge calibration
forecast work and all three count fits plus subset fit per arm. Both heldout
arms compute calibration (cold discards weights), preserving matched fit work.

Publication invalidates outstanding old-model selector advice. Anti-Pigeon can
independently split on baseline evidence; it does not assert the holdout weights
are statistically calibrated. Both delayed schedules and missing labels follow
the previous forest-delay protocol exactly. All512 original forecasts are
scored, including censored labels; calibration never rewrites earlier scores.

Primary: warm delayed post-change gain>=.005 with paired lower(mean-3.5SE)>0
against BOTH unchanged and cold, for each of four changed cases in each cohort
(16 gain comparisons). All six cases, both schedules, both cohorts: full and
post gain lower>=-.01 versus EACH control (48 non-harm comparisons). Approximate
fixed-sample screens, not confidence sequences. Cold ablation performance also
reported. No holdout-size or prior tuning after collection.

Before collection verify unchanged-control exact immediate/delayed forecasts,
fit schedule, partition separation, initial weights reaching first served
forecast, and model label-isolation tests. Full replay and independent score,
availability and calibration-weight reconstruction required. Keep failed gates.
Measure complete fixture performance after quality runs finish. No production,
paper, commit, push or private-data access. No broader-goal completion claim.
