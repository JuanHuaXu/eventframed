# Origin timing diagnosis v105

Consumed-data diagnosis only, not confirmation. Trace all128 v104 delayed
switch trajectories: both phases, both directions, all32 indices. Do not select
only adverse individual runs. Require the complete original record, including
all three arms, to remain bit-identical.

Trace every selector update in the role-carry arm: issued origin, actual arrival,
prefix release clock, version crossing, original raw forecast, outcome, bank
weights before/after and additional waiting beyond arrival. Count fully applied
and bank-only updates; exclude censored labels from update traces.

At each scored release clock, use the currently published fitted models and
current simulator regime to form a diagnostic full-input raw-bank Brier profile.
This is a uniform average over all512 inputs, not the acquired-view scored law,
not a routed-gate risk, and not a counterfactual rerun. It is used read-only:
an instantaneous increase indicates that the selector update moved the current
raw mixture toward a worse forecast under this diagnostic. It does not prove
the cause or size of the served-law regression. No oracle value feeds learning.
Flush updates after clock255 have no current-forecast risk annotation.

Report counts, mean age and mean additional waiting, and the signed diagnostic
risk changes. Split post-change updates into pre-change and post-change origins;
also split bank-only versus current-version updates. Preserve the acquired
masks in the complete parent record, but do not attribute observation-policy
effects to raw-bank weights alone. Compare profiles and risk calculations with
independent reconstruction. Replay the traces. Do not use these consumed results
to claim a tuned half-life or a validated new selector.
