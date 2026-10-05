# Prospective Research Baseline Policy

Selection uses the complete consumed 40-world, 120-cell native cohort, with the
original 400-ms whole-synthetic-core ceiling. Expected issued Brier is minimized
among comparable complete candidates within that ceiling. Exact loss ties prefer
lower mean core time. Different cohorts, missing cells, nonfinite metrics, and
prepared single-operation timings are not eligible substitutes.

The recorded V72 cohort contains 18 same-cohort arms. Adaptive has the lowest
expected issued Brier, 0.2147478547980669, with recorded mean/worst core
201.944/211.952 ms. Full remains the speed control at loss 0.2246990964333006 and
50.307/68.356 ms. Some weaker models are faster still; no one arm is claimed to
be globally best on both axes. V75's best mean model has loss 0.21958154845158395
and exceeds the core ceiling in all cells. V83's prepared query timings cannot
enter this whole-cohort ranking because it lacks a comparable quality result.

Before finalizing the registry, rerun the existing Full and Adaptive control
functions on all 40 worlds and three schedules using consumed seed 2026105407.
Do not open reserved seeds 2026105409 or 2026105411. Preserve all 240 arms and
independently recompute their 576,000 issued expected losses. Aggregate risks
must reproduce the recorded control risks within 1e-12. Retain every fresh cost
measurement. Also rerun the existing future-prefix test and selector boundary
tests. This is replication of consumed development evidence, not confirmation
or a significance test for universal superiority.

PrimaryArm=adaptive applies to the prospective research registry and its new
native-control audit. Full and all historical baselines remain comparison arms;
historical runs and frozen protocols are not modified. Daemon defaults are not
changed. V83 is the current numerically repaired research implementation, not
an accuracy-promoted deployed baseline. All seven whole goals remain open.
