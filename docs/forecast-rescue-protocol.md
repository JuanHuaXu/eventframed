# Composed forecast rescue protocol, 2026-09-06

Confirmed: grid v1 loses on stationary .20/.80, and its better raw belief loses
after .9*.5+.1*belief composition on gradual drift. Fixed reset noise explains
the former; shrinkage toward a mismatched baseline explains the latter. These
are algorithmic/model limitations, not malformed data. Prior confirmation is
now design knowledge, never reused as untouched rescue confirmation.

Rescue: four complete Bernoulli forecast experts, in order: calibrated baseline,
existing calibrated grid blend, raw grid predictive, raw Beta mean. Convex
mixture with prior [.7,.1,.1,.1], fixed share .002 per eligible feedback, log-loss
Bayesian weight update, evidence weight clamped to [0,1], floor 1e-6. No parameter
search. This is modular expert aggregation, not a posterior truth probability.
Do not force an estimator of true usefulness to compensate for shrinkage;
instead allow the complete forecast to learn when shrinkage is unhelpful.

Freeze before execution: design seed 2026090611, untouched confirmation
2026090612. 64 trajectories x1000 observations in each of the original eight
scenarios, crossed with baseline .2/.5/.8 and alternating .2/.8 every 50 steps.
Predict before outcome. Controls: old two-hypothesis composed law, grid composed
law, direct Beta. Report all trajectory Brier/log loss; signed gain = control
minus rescue. Approximate trajectory-normal simultaneous intervals use 3.6 SE
over 32 scenarios x2 controls x2 metrics (128 comparisons). Pointwise 1.96 SE
also reported. Not exact/sequential coverage.

Primary rescue: at baseline .5, stationary .20/.80 and gradual drift must have
positive simultaneous Brier gain over grid. Guard: no scenario's simultaneous
upper bound on mean rescue excess Brier over old may exceed .003. Keep all
failures. These are forecast-level tests, not retrieval/answer-quality tests.

Integration if supported: explicit opt-in flag requiring authenticated grid
mode; no contextual/hierarchical scoring; residual mode disabled for the first
rescue so the evaluated expert mixture is the final scored law, not a detached
intermediate. Existing rank correction is unchanged. Persist as-of expert
predictions with the journal, update selector in the outcome transaction, and
never recompute historical experts using the revealing outcome. No selection,
trust, or Anti-Pigeon bypass. Stale-epoch/policy feedback cannot train the new
selector. Splits discard pooled selector state; resets start at its prior.
All storage is fixed-size; default behavior stays unchanged.

Bug hunt: immutable prediction values, read purity, invalid-state fallback,
replay/restart, late feedback, policy transition, reset/split and score wiring.
Full Go tests, race tests, vet/build. Paired authenticated two/grid/rescue disk
benchmarks using existing 50-event fixture, 1/4 workers, 500 ops x3; report p99
and throughput separately. Do not test or modify production.
