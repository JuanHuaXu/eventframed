# Read-only split mediation diagnostic

Replay all160 archived subset-gate-pair trajectories; no new sample or retuning.
Before each outcome, capture both actual subset forecasts, split status, observer
guide, effective long-slot weight, pooled/local forecasts on the actual observed
mask, and the emitted mixture with only that slot exchanged. Nil local/pooled
models are unavailable, not fictitious0.5 counterfactuals. Determine the effective
weight through the existing Forecast API and check its affine/clipped identity.
No diagnostic value enters prediction, evidence, acquisition or model updates.

Every archived original metric/tape and both candidate metrics/tapes must match,
as must fit seed/hash/count. Capture current diagnostic sources and original
archive hash. Save per-frame numbers for independent arithmetic checks, including
the outcome only after prediction. Analyze fixed64-step windows and the period
when only the newer gate has split. Report direct fixed-mask slot substitution
separately from observed between-arm gain; it is not a full alternate policy.

This consumed diagnostic cannot validate new quality, coverage, causality or
performance claims. Do not turn hindsight action selection into a predictor.
