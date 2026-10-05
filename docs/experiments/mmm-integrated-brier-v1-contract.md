# Proper-loss-aligned outer selector pilot

Distinct from the earlier failed head-bank strong-Brier study: test the retained
subset workflow's four-expert outer selector, with observations held to its
unchanged inherited-weight mixture-gate control. No new hidden features, fits,
labels or acquisition budget. Keep inner subset/count learning, Anti-Pigeon
authorization and conservative pooled-to-local action unchanged. No birth prior.

Compare original log-loss update/linear forecast, scalar-Brier eta2 update/linear
forecast, and the SAME eta2 weights with strong binary Brier substitution from
Vovk/Zhdanov2009 Algorithm1. Their summed two-class loss is twice scalar Brier.
For normalized predictive weights, z_y=sum w_j exp(-2(p_j-y)^2), then
p=.5+(log z_1-log z_0)/4. Keep current expert clipping, prior(.7,.1,.1,.1),
fixed sharing.002 and gate action. Do not inherit the paper's static cumulative
bound after these sharing/revocation adaptations. Updates use issued experts.

Fresh seeds2026092131/2026092132; five scenarios,16 independently fitted4096-label
bases per cohort/scenario,512 live frames. Primary strong versus original:
.005 mean post-gain plus positive mean-minus3.5SE, .01 full/post non-harm. Retain
linear-eta2 comparison to isolate substitution, no eta/prior/threshold sweep.
This is exploratory, not rare-revocation or delayed/dependent-data validation.

Verify scalar formula against literal two-outcome substitution, one-step loss
inequalities and matched weight histories. Require unchanged control tapes,
split timing, fit counts and exact acquired costs. Capture sources and all
results. Test-only code, no production, paper, dependency or remote changes.
