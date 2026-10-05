# Arrival-learning dependence breadth

Use all12 existing forestDependenceCases,16 trajectories per original cohort,
seeds2026092231/2026092331, role4 training seeds. These allocations are consumed
by the earlier forest-dependence study; do not call either cohort fresh.
Run unchanged innerArrivalRun under immediate and jitter0..31/missing.2,
four arms original/outer-only/full/fixed-view. No parameter tuning or omitted case.

Report full/post Brier, correctness, acquisition cost, split times, and all
fit-origin sets. Immediate arms must agree. Every training origin must be an
arrived nonmissing audit; predictive output precedes same-tick labels.
Compare full versus original and full versus outer-only. Descriptive screens
retain .005 mean gain and positive mean-3.5SE on changed delayed cases, and
.01 upper-harm allowance (mean+3.5SE) for all full/post cells. These are consumed
fixed-sample diagnostics, not research-wide coverage or new adoption criteria.
All actual coordinates (foreground, audit, monitor) remain accounted separately;
do not infer equal cost solely from a common per-frame cap. Goal7 is not tested.
