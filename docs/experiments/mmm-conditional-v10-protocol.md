# Conditional integration v10 diagnostic

Frozen before calculation. Replay v9's already-consumed records; this is a
paired mechanism diagnostic, not an untouched confirmation or a policy rescue.
Keep each tree, acquired mask, input, label and audit cadence fixed. Use the
adaptive-retained arm's mask, with no new observations. Score tree forecasts
only after the first supported fit. Compare:

1. Existing uniform missing-bit integration, matching the recorded tree forecast.
2. Joint input frequencies from the last <=256 admitted audits, plus one total
   uniform pseudo-observation over512 full assignments.
3. Oracle input distribution from the declared generator, diagnostic only.

For fixed fitted tree f and observed x_M=v, forecast is
sum_{x:x_M=v} w(x)f(x) / sum_{x:x_M=v} w(x).
Weights use inputs only, not outcome agreement. Estimate weights at the existing
32-audit first fit and each16 thereafter, after that frame's prediction. Current
or future inputs cannot enter earlier estimates. Tree remains fitted to last64
audits with the v9 seed, unchanged. Oracle is not a deployable result or a
guaranteed upper bound for a misspecified outcome predictor.

Precompute a frozen table over3^9 partial assignments. This is bounded research
code; exponential scaling to arbitrary feature counts is not proposed. Forecast
lookup is O(9); fitting enumerates512 full inputs plus3^9 partial assignments.

Report paired full/post mean Brier by generator/scenario/split, using per-stream
means to avoid unequal available-count weighting. Every result remains visible.
Advance empirical integration to a new online-policy trial only if clustered
shift128 post improvement >=.005 in both splits and no group/window mean harm
>.01. These finite screens neither establish calibration nor replace v9 gates.
Oracle comparison separates estimation error from integration's limited value.

Tests: direct weighted enumeration over every partial assignment, uniform
parity, full-observation invariance, immutable snapshots, invalid-weight rejection,
and exact match to every recorded uniform tree forecast. Audit source hashes and
retain raw diagnostic forecasts. No serving changes or production enablement.
