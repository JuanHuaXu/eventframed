# Frozen full-input learner comparison v118

No quality output inspected before freezing this protocol and its verifier.
This tests learner quality at equal evidence, not partial-view MMM integration.
Keep every prior failure and existing specialist. No penalty tuning or sample
extension after results. Solver failure aborts the experiment, not a dropped row.

## Data and policies

Use all9 new transfer family/mode cells plus all12 original Boolean regimes,
both phases and32 indices each. Every latent stream has immediate and delayed/
missing schedules:1344 latent trajectories,2688 runs,256 scored ticks per run.
The transfer base is2184111800, with the unchanged generator seed offsets.
The Boolean base is2188111900 +phase*1000000 +case*10000 +index*10, RNG roles0..4.
Audit effective seeds against all archived blocks, v116 and consumed QA.

Keep16 initial examples, as-of64/32 fits every32 ticks, delays0..31, missing.2,
and the unchanged truth/input constructions. Earlier eligible feedback arrives
before fitting, current zero-delay labels only after all forecasts. Each query
exposes all9 coordinates equally to every learner. Initial evidence costs144
coordinate reads; query evidence costs2304 per trajectory. Arriving labels do
not charge those same already-observed coordinates again. This differs from the
partial-view fitting-audit budget and must not be compared as equal acquisition.

Eight independent full-input arms:
0 generic64,1 Boolean64,2 generic32,3 Boolean32,
4 ridge64,5 ridge32,6 context-tree64,7 context-tree32.

Freeze all existing fitters. Ridge uses exactly the contract in
research/ridge-challenger-component.md: summed loss, penalty1 including intercept,
32 Newton steps,24 backtracks, gradient tolerance1e-8, forecast floor1e-12.
No teacher identity/relevance, query outcome, future label or changepoint enters
fitting. The existing context-tree prior/depth stays unchanged. This does not
claim a logistic model can represent parity or arbitrary interaction tables.

## Fixed screens and interpretation

Score expected Brier over all256 and terminal ticks192..255. Report expected
accuracy/log loss, realized scores, full-input teacher floor and ridge fit
diagnostics separately. Use paired mean +/-3.5SE over32 trajectories, with the
same .01 non-harm tolerance and .005 mean-gain floor plus lower bound >0.
These are approximate fixed-sample screens, not anytime or research-history-wide
guarantees. Two schedules of one trajectory are not independent samples.

For each candidate4..7 require:

-336 non-harm screens: against its matched-window generic AND Boolean controls,
 in21 cases*2 phases*2 schedules*2 segments.
-32 terminal recovery gains against matched-window generic: all6 changed
 transfer cases plus both original Boolean switches, both phases/schedules.
-4 stationary all-frame gains against matched generic: additive for ridge,
 hierarchy for context-tree, both phases/schedules.

Thus372 requirements per candidate,1488 total (1344 non-harm,144 gain). Report
every candidate's broad standalone verdict, not just a favorable family. In
addition report its12 predeclared structural-target gains: stationary plus both
changed additive modes for ridge, or hierarchy modes for context-tree. Those12
are a subset of existing gain screens, not extra gates or a substitute for the
broad verdict. A structural-target pass only warrants a composite hypothesis;
it does not justify removing an incumbent or claiming integrated rescue.

## Reproducibility and boundaries

Record all packets, forecasts, teacher identity/rules, eligible fit origins,
ridge coefficients/convergence stats and source hashes in an exclusive0600
artifact. Independently reconstruct scores, teacher probabilities, as-of fits,
seed uniqueness, pairing and gate counts. Require exact full replay and unchanged
hashes. Freeze the generator/solver/verifier/protocol before generation.

Generation/replay durations are not serving latency. Retain the existing component
benchmark and no production/OpenClaw, private-data, deployment, pushing or
whitepaper promotion. All seven directions remain open unless their full scope
is separately established. A standalone learner failure is not automatically a
failure of a future incumbent-preserving composite, which needs its own test.
