# Frozen eight-index breadth screen

Expand cadence and feedback/budget-v1 unchanged to source indices 0-7,
all 21 scenarios, both phases and arrival schedules at clocks 128-159.
Expected 672 records, 21,504 forecasts. Keep 64-origin window, pi=1/255,
unit slab/intercept variance, 1024-iteration cap, convergence and integration
budgets, equal expert prior, learning rate 1 and .01-per-forecast loss budget.
No parameter selection or case exclusions based on outcomes.

All source trajectories were previously consumed in other model experiments;
index 0 additionally selected this research lead. Report it separately from
indices 1-7. Neither set is untouched confirmation. Preserve every failure.
Use existing frozen32 screen at clock128 as the source-aligned fit reference.
Check every dynamic fit origin, moments, convergence and every budget prefix.
Index0 results must exactly reproduce the earlier pilot. No silent retries
with a different model if any fit hits its cap or fails.

Report per-record and per-scenario harm, pooled and phase/schedule scores,
fit count and collection cost. Report paired trajectory uncertainty for
scenario means via deterministic trajectory bootstrap, explicitly exploratory
and pointwise, not simultaneous safety certification. A .01 expected-score
harm screen is diagnostic; realized-loss budget compliance alone is not the
goal. Broader clocks and untouched confirmation remain subsequent requirements.
