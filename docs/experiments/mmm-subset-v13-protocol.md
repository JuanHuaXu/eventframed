# Bayesian subset challenger v13

Frozen before experiments. This is our finite-domain research adaptation, not
CTW, Bayesian CART search, or an inherited generalization guarantee. See
research/probabilistic-challenger-lead.md for primary-source motivation.

Enumerate all512 feature subsets S of nine binary coordinates. Conditional on
S, each outcome cell has independent Beta(1/2,1/2) prior. Prior probability of
S is (1/3)^|S| (2/3)^(9-|S|), fixed before evaluation. For admitted training
inputs treated as covariates, evidence is product over cells of
B(y+1/2,n-y+1/2)/B(1/2,1/2), WITHOUT binomial coefficients. Average each subset's
posterior predictive under normalized prior*evidence. Compute in log space.
This models conditional labels, not causal truth or a posterior over input law.

Fit on last64 admitted audit labels, same32-first/16-refit cadence. Partial
forecasts and acquisition use the v11 conditional table with joint frequencies
from last<=256 admitted inputs plus one total uniform pseudo-observation.
The product of this frozen input estimate and the subset conditional law defines
the working joint model. This plug-in input treatment is not fully Bayesian.

Two modes: original uniform forest and subset challenger. Preserve all four
outer arms (fixed count, replacement, adaptive retention, static retention).
Only challenger fitting/forecast/acquisition differ. No extra labels or live
coordinates. Keep incumbent and short/long count models, priors, delays and
missing-label handling unchanged. Mode strings in artifacts: forest/subset.

Six scenarios: stable05, shift128, recurring, delayed_missing, interaction, null
(indices0,2,6,7,8,9). Three generators fair/biased/clustered; two fits*four streams
per fit*two splits*two modes =576 streams of512. Fit2026106201,
design2026106202, confirmation2026106203; existing role encoding. Matched fitted
incumbents reused across modes and splits; no tuning between splits.

For each retained subset arm: mean full/post Brier harm<=.01 versus fixed count
and matching retained forest arm in every group; shift128 post gain>=.005 versus
fixed in every generator/split; clustered shift128 post gain>=.005 versus matching
forest in both splits. Retain every failure. Pilot means are not population
confidence bounds. Passing permits a stronger test, not production promotion.

Verify evidence versus closed-form Beta integrals, order invariance, normalized
weights, null/invalid data, exact forest control, pairing and leakage guards.
Archive full sources, raw forecasts and labels. Measure standalone fitting and
lookup separately from serving. Enumeration is exponential in nine bits and
must not be advertised as scalable to arbitrary event representations.
