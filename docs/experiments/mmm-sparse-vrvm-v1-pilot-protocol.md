# Sparse variational as-of pilot

Freeze before collection: both phases, all21 scenarios, both schedules,
index0 only; publication clocks0,128,224; latest eligible64 and32 labels.
Exactly504 fits and16128 next-window forecasts, with integrated and plug-in
mean predictions from the SAME fitted Gaussian. All15 original control
forecasts are retained. Every fit is charged; no approximation/stop is dropped.

Only initial16 labels and earlier nonmissing labels arriving by publication
may fit the model. Freeze the fit for the next32 forecasts. Q and future Y
are evaluator fields written after forecast computation. No teacher, case,
seed or current/future label enters fitting. Store the exact Gaussian-input
xi, inverse expected precisions, mean and diagonal for independent audit.
Final xi is different from Gaussian-input xi at an iteration cap: never
substitute it during reconstruction.

Fit/integration settings are the existing frozen contracts, including64-step
cap and no pruning. No confidence intervals or broad pass/fail quality gates
can be estimated from one trajectory per cell. This is a consumed feasibility
pilot, not the full2688-record experiment or a fresh confirmation. Any later
quality test must retain the original controls and whole success criteria.

Before collection, test unavailable-label and evaluator-metadata poisoning,
fit-origin equality and reconstruction from the recorded Gaussian input state.
Preserve all failed attempts. Full pilot audit/replay follows if still within
the user's weekly allowance stop boundary; clearly record pending work.
