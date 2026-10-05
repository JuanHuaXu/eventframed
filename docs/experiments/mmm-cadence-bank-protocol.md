# Cadence bank diagnostic

Analyze both the original32-clock and new8-clock four-expert forecast banks
on every consumed v120 trajectory. No learner, gate, threshold or prediction
is changed. The previously audited faster-cadence artifact is the input.

For each frame, the Bernoulli Brier floor is Q(1-Q). Project Q onto the interval
from minimum to maximum issued expert probability to obtain the convex-hull
oracle. Its Brier minus the floor is bank gap: no convex weighting of these
particular issued predictions can remove it. Served Brier minus hull Brier is
oracle selection gap, not necessarily achievable with available observations.
Check the additive identity and nonnegativity rather than assuming either.

Also compute best fixed expert, at most one/two switches, and unrestricted
framewise expert with the existing tested switch-budget dynamic program.
Hindsight chooses the initial expert and switch clocks using Q. Compute full256
and terminal64 independently, permitting a fresh oracle start in each window.
Compare only the SAME four expert roles across cadences; the earlier six-expert
oracle included additional predictors and is not this control.

Report every phase/case/schedule/window with paired mean +/-3.5SE over32
trajectories, explicitly exploratory, not simultaneous coverage. Independently
match loss totals to the cadence summary, hashes to the source artifacts and
all record identities. Preserve each trajectory and run scoring replay.
Large oracle selection gap does not prove learnability; large bank gap does
not identify a unique needed model or prove that all possible refits fail.
This diagnostic closes none of the seven research goals by itself.
