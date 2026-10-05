# v96 interval-local weighting (frozen before outcomes)

Same learned-model and observation protocol as v95: two phases,32 paired
streams per case,12 cases,16 initial full-frame labels,256 scored steps,
refit every32 steps from the latest64 received frames. Both full and mask63
views have the same full training evidence. No extra model fits or acquisition.
Seed = 2050119600 + phase*1000000 + case*10000 + index*10; roles0/1/2 remain
rule/input/outcome. Disjoint phase rule pools and case definitions are v93's.

Retain all five v95 controls, adding one primary interval-weighted arm.
Active geometric intervals are [i*2^k,(i+1)*2^k-1], i>=1, covering rounds1..256.
Instantiate the bounded kernel with horizon256. Every interval runs rho0 Brier
aggregation with generic prior.95 and rate.5 over the same two learned forecasts.
Meta entry weight and rate are min(.5,1/sqrt(interval length)); after outcomes,
multiply each weight by 1+rate*(weighted-average interval loss - its own loss).
Output the convex mixture of interval forecasts, not a sampled expert.
Weights restart only at predetermined interval boundaries, not model refits
or known regime changes. Separate full/mask63 meta states use the same labels.

This is a deterministic finite-horizon SAOL specialization with a surrogate
loss update. It does NOT inherit the v94 constant lifetime bound or a1% safety
guarantee. Verify against an explicit interval enumeration with direct base
probability updates, and check convexity and lifecycle errors. The published
worst-case constants are too loose to establish our short-window thresholds.
The bound checks stored in records apply only to the retained v95 online arms.

Keep v95's58 quality gates for the new primary arm:48 all-stream non-harm,
6 stationary interaction gains,4 late-half recovery gains. Add48 late-half
non-harm gates covering all cases/phases/views, so past gains cannot hide late
harm. Total106 gates, all required; paired32-trajectory z=3.5 normal intervals,
harm upper<=.01, gain mean>=.005 and lower>0 as in v95. Approximate simultaneous
screening, not anytime confidence sequences. No parameter sweep.

Full768-record replay, hash-verified independent summary, race smoke/reference,
vet, same-fixture full-stream benchmark and separate256-update kernel timing.
No production, delayed-feedback, omitted-label or real-agent claim.
