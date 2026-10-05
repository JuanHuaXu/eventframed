# Delayed switch V43: computational rescue results

2026-10-03 local. Four separately frozen screens preserve the failed attempts
and the eventual preliminary computational PASS. No quality adoption follows.

| Implementation | Immediate ms (3 runs) | Fixed150 ms | Fixed299 ms | Constructor max B |400ms/8MiB |
| --- | --- | --- | --- | ---: | --- |
| initial log filter |354.05-357.60|413.50-418.25|466.56-469.22|7,408,584|FAIL work|
| cached constants |340.07-347.76|396.03-398.67|444.11-447.40|7,403,232|FAIL work|
| log-only cache |344.53-346.89|389.51-390.59|432.74-441.75|7,247,696|FAIL work|
| unit-emission spans |340.31-344.58|343.22-354.42|344.93-351.14|7,268,224|PASS|

Artifacts are `research/switch-v43-cost-{initial,cached-constants,log-only,unit-spans}/`.
Each records sources, runtime/host/load, command, terminal race/vet and all18
benchmark rows. All frozen V41/prior learner bytes remain unchanged. These are
three fixed computational repetitions per case, not statistical tail-latency
estimates, a randomized A/B performance study or loaded serving certification.

All Full/Adaptive/rich-moment2 construction, issue, visible-label updates,
mixture replay, round snapshots and final drain are timed. Deterministic labels
are a stress fixture, NOT stochastic quality outcomes. Acquisition, scoring,
serialization, disk, network and served-packet time remain outside this screen.
No inference of equal TOTAL cost or continuous-learning completion.

Initial supplemental CPU profiling of the delayed loop attributes0.25s of2.08s
sampled cumulative CPU to `transitionLogs` and0.43s to mixer Resolve. Global
math.log samples also include child learners, so cannot all be assigned to the
mixer. That profile motivates measurement; it is not an independent confirmation.

Constant caching has exact-bit transition regression over endpoint/interior/
subnormal hazards and epoch replacement. Removing the linear diagnostic cache
does not change filtering. The span rescue uses the explicit T^k identity in
the prospective span contract. It passes twelve root race tests, independent
small full-history/path checks and ALL4096 prefixes after sparse out-of-order
oldest/newest/interior reveals, cancelled spans, tiny/zero/full hazards,
atomic failed insertion and epoch replacement. Existing3e-12 tolerance remains.

The4096-position oldest-arrival microbenchmark drops from2.06-2.74ms to
0.792-1.500us because its4095 later positions are ALL unknown, not4095 known
emissions. Do not generalize that speedup to dense resolved suffixes: worst
case O(N*T) remains. The current span index adds O(T) bounded insertion work
and O(log T) predecessor lookup. Constructor storage remains below8MiB.
The microbenchmark returns the ORIGINAL issued forecast, not a newly projected
tail forecast; lazy tail projection is charged in the whole-pool snapshots and
subsequent issues. The microbenchmark alone is not resolution-plus-serving cost.

Next freeze NEW broad-domain quality cohorts and static/two-share controls.
All seven WHOLE goals OPEN/ACTIVE; no production, private-corpus or paper edits.
