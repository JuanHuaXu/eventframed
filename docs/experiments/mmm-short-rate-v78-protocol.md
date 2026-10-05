# Short-window rate model v78

Frozen before fresh data. Research-only response to v77's measured rate lag.
Keep the v76 32-observation allocation/augmentation model, query budget, actual
propensities, eight starts, factor floor .08 and threshold100 unchanged.
Only the rate-model probabilities use the last eight observed outcomes per
channel, with Dirichlet counts (.5,1,.5). No generator identities, future
observations or known change times enter this model. Eight is frozen as a
fourfold reduction of the measured lag horizon, not selected from a sweep.
The predictive rate maximizes expected log-factor by the same 24 bisections
and cap as v76, evaluated with the ORIGINAL allocation/augmentation snapshot.
Do not change the query probabilities to match the short rate model.

Four paired arms: uniform, fixed augmented, long-rate v76, short-rate v78.
Fresh bases2026117801/02, seed=base*1e6+scenario*1000+stream. Ten unchanged
scenarios,512 steps,512 streams per cell, two phases,10240 total streams.
All arms share the same counterfactual random tape; rate arms share selected
observations. Controls run after the short candidate and cannot inform it.

Candidate gates remain: primary sparse128/256 restricted-mean improvement at
least10% vs uniform, paired z3.3 lower gain>0 and no extra premature alarms;
each null Wilson95 upper<=.02; other alternatives at most10 steps mean delay
harm. Six alternatives require excess misses<=.01 and simultaneous v73 paired
upper<=.02, alpha=.05/12 across alternatives and phases. No tuning between
phases. Report phase results separately, and any failed protection cell rejects
adoption. Passing relative weak/late guards is not absolute adequate detection.

Verify the32-window special case against v76, chronological ring wrapping,
history immutability, sign/coordinate equivariance, optimizer grid, factor
bounds, invalid-window rejection and exact original query parity. Freeze code,
protocol/evaluator hashes, retain per-stream alarms and tape hashes, run full
replay and race/vet. Measure balanced and sustained-positive rate/gate loops
separately; neither is a request-latency benchmark. Do not enable serving code.
