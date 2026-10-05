# Sparse variational convergence diagnostic: frozen before collection

Question: does the 64-step iteration cap explain the capped model's weak
pilot forecasts? This is an optimizer diagnostic, not a changed prior or model.

Select index 0, both phases, all 21 scenarios, both feedback schedules,
publication clock 128, and latest eligible 64 labels: exactly 84 fits and
2,688 next-window forecasts. This subset spans the pilot population while
holding window size and observation clock fixed. It cannot validate the
omitted clocks, the 32-label window, or full trajectory quality criteria.

Use the identical frozen 256-feature model, prior, initialization, coordinate
updates, bound, stopping tolerance 1e-6, and integration rule. Raise only the
iteration cap to 1,024. Start fresh, not from a serialized rounded state. Verify
each extended trace starts with the exact 64-step trace in the original pilot.
Retain all original forecast controls and the matched capped candidate.
Keep the same origin and as-of exclusions; no evaluator information may fit
the model or choose its iteration budget. Do not discard capped/failed fits.

Report convergence fraction, final bound increments, matched objective change,
integrated and plug-in Brier changes, all controls, and total collection cost.
A converged label means only the existing numerical stopping criterion; it
does not imply a global optimum. Monotonic objective improvement alone is
not a predictive-quality win. All-cell prediction comparisons are descriptive,
not fresh confirmation or confidence intervals from one trajectory per cell.

Check the default 64-step wrapper remains equivalent, invalid budgets reject,
extended prefixes match, bound monotonicity holds, and future/evaluator labels
do not alter the extended fit. Independently reconstruct the final Gaussian
and probabilities as in the original pilot. Require full byte-identical replay.
Do not launch a larger-cap grid after seeing the results; reassess the source
model and numerical behavior before proposing another fitting change.
