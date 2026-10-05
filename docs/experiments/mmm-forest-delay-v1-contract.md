# Frozen forest delayed-quality screen

Compare unchanged uniform-input and held-out-forest learners, each with its own
observer. Six cases: perfect-copy bit/XOR2 shifts, noisy-copy bit shift, reversed
dependency XOR2 shift, stable noisy-copy bit, and null uniform. Keep .05 outcome
noise except the null case. 16 independently fitted4096-label bases per case,
two cohorts2027010151/2027010251,512 predictions. Each latent trajectory runs
immediate and delayed schedules; these schedules are paired, not independent.
Total192 independent trajectories,384 schedule-runs,768 arm-runs.

Delay is uniform0..31, missingness.2, jointly applied to live/reference outcomes,
independent of outcomes and input. Separate role7 RNG; roles0..6 unchanged.
Journal capacity64, expiry age32; origin-ordered selector updates; old model
generation advice discarded. Model publication and external splits invalidate
old selector advice. No runtime sees the simulator change clock.

The fixed-baseline Anti-Pigeon investigator processes available pairs in resolved
origin order; missing pairs are skipped, not negative. It has independent
authority to split even when selector advice is stale. At each clock: issue
forecast, acquire audit inputs, reveal eligible labels, settle selector prefix,
expire missing prefix, process gate prefix, then publish fits. At most one fit
per clock if the arrived-audit count crosses a32+16k threshold. Fit the last64
and256 arrived audits ordered by origin; pooled fit uses last128 of each stream.
Full audit input access costs18 coordinates, including later-missing audits.

Precollection compatibility must match earlier immediate Full/Post metrics and
split times. Verify deterministic delayed replay, identical latent tapes and
fit-origin availability. Score all512 forecasts including missing-label origins.
Record per-frame original forecasts, fit origins, settlement, monitor/audit cost.

Primary delayed changed-case gain>=.005 and paired lower(mean-3.5SE)>0 in each
cohort. All cases/schedules full and post lower gains>=-.01. Intervals are
exploratory fixed-sample screens, not confidence sequences or a guarantee over
research history. Preserve all failures. No threshold or forest tuning.

Freshness here concerns synthetic evidence clocks, not persistence or serving.
Total costs include monitor/audit acquisition and fits, separately from foreground
acquisition. Wall-clock collection time is not request latency. No production,
whitepaper or remote edits, and no full-goal completion from this study alone.
