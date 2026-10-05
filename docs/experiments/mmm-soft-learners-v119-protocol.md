# Frozen uncertainty comparison v119

Predeclared before inspecting fresh quality output. Retain all v118 cases,
controls, evidence budgets, clocks and harm/gain thresholds. No quality-based
penalty, convergence, quadrature, window or sample-size changes after generation.
This tests full-input learner quality, not partial-observation integration.

## Inputs and arms

Use the unchanged v118 generator over9 transfer cases and12 Boolean regimes,
2 phases*32 indices*2 schedules:2688 runs,1344 latent trajectories,256 ticks.
Transfer base2192112000; Boolean base2196112100. Preserve role and case offsets,
audit all6720 effective seeds against archives including v118. Each pair shares
inputs/outcomes; schedules are not independent trajectories.

Keep16 initial labels, as-of64/32 fits every32 ticks, delay0..31 and missing.2.
All nine input coordinates are visible to every arm:2448 total coordinate reads
per trajectory including initial examples. Current labels follow all forecasts.
There is no teacher identity, future evidence or outcome in model fitting.

Arms0..7 are unchanged generic64,Boolean64,generic32,Boolean32,ridge64,ridge32,
context-tree64,context-tree32. Add8 variational64 and9 variational32 following
research/variational-logistic-component-contract.md. There is no learned mixture
or incumbent suppression here. Preserve generic and nonlinear specialists.

## Requirements and reporting

Expected Brier all256 and terminal64, expected accuracy/log loss, realized losses,
teacher floor and all per-window moments/diagnostics are recorded. Intervals are
paired mean +/-3.5SE across32 trajectories; approximate fixed-sample screens,
not anytime, simultaneous population certificates or research-history guarantees.
Non-harm requires upper<=.01; gain requires mean>=.005 and lower>0.

Retain all1488 v118 screens for candidates4..7. For each variational candidate,
apply the same372 standalone requirements against matched generic/Boolean:
336 non-harm,32 changed terminal gains,4 stationary additive gains. Its12 additive
structural targets remain a subset, not a substitute for broad validation.

Additionally, for each variational candidate require:
-168 non-harm against its same-window MAP control over all21 cases, both phases,
 both schedules and both segments.
-40 gains against same-window MAP:32 terminal gains over all6 changed transfer
 cases and2 Boolean switches, plus4 stationary additive and4 null all-frame gains.

Six candidates*336=2016 standalone non-harm; six*36=216 standalone gains;
two*168=336 MAP non-harm; two*40=80 MAP gains. Thus2352 non-harm plus296 gains
=2648 total. Each variational candidate has580 requirements. Separately report
the208 MAP-replacement tests and372 broad standalone tests; neither passes by
averaging failures away. A narrow specialist gain only motivates a subsequent
independently tested composition, not a completed rescue or seven-goal closure.

## Audits

Exclusive0600 artifact records initial packets, all forecasts, all eligible fit
origins, MAP coefficients, variational mean/covariance/bound/residual/iterations,
and source hashes. Require exact full replay, independent teacher and score
reconstruction, eligible origin checks, pairing and seed checks. Independently
reconstruct variational fixed-point equations, positive covariance and bound,
and issued forecasts by wider-domain Simpson integration using tanh, not by
calling the implemented predictive function. No dropped failures or silent
fallbacks. Computation duration is not serving latency. No production access,
deployment, pushing or whitepaper promotion.
