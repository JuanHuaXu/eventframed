# Frozen multi-clock screen

Before collection: retain the complete pilot's model and numerical contract
unchanged (pi=1/255, slab variance1, intercept variance1,255 masks,64-label
window,1,024 fit cycles, actual mixture integration). No model selection.

Source is the consumed soft-learners-v120 tape; all21 cases, both phases,
both evidence schedules, indices0 through7, clocks0/128/224. Total2,016 fits
and64,512 forecasts. Retain32 forecasts per publication; never fill gaps
between these publication blocks with inferred results. Initial clock has
only16 admitted labels. Every source-origin set must match that publication's
stored64-window origin list. Fitting receives only admitted bits/outcomes.
Errors stop the run; capped fits remain in output. Use exclusive new files.

Purpose: test reproducibility of the single-index pattern, distinguish
initial/stationary/midstream/late behavior, and decide whether full32-index,
all-publication testing is justified. This is screening on a consumed tape,
NOT independent confirmation, a full adaptation-delay analysis, or the full
non-harm/recovery gate. No prior changes after viewing this screen.

Report every phase/schedule/clock cell, expected and realized Brier for
mixture, mean plug-in, generic64, Boolean64 and Markov. Retain per-case and
per-index paired results so improvements are not hidden by overall averages.
Check all dimensions, admitted origins, controls, convergence and moments
independently; require replay equality. Compare matching index0/clock128
records against the completed pilot. Report total collection cost separately
from per-fit/per-prediction benchmarks. Keep negative and inconclusive cells.
