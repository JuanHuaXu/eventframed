# Provenance acquisition versus scoring v5

Frozen diagnostic protocol before execution. Preserve all v3/v4 failures.

## Design

Cross fixed/informed Gini acquisition (A) with fixed/informed posterior scoring
(S): FF, FI, IF, II, where the first letter denotes acquisition. Both scorers
under one acquisition policy receive exactly the same chosen pair and outcome.
Selection uses only its designated acquisition model's observed history; the
other scorer cannot change the next query. Both policies share the original
outcome tape and eight provenance signals. No actual modes or target label enter
any selector or posterior update.

Reuse v4's frozen disjoint calibration without retuning. Sixteen observations
plus eight charged signal checks for every arm; 24 units per arm. All source
slots unique. Forecast before each observation and after all16, with no stopping.

Cases: independent20, copied20, mixed20, matched05, matched20, matched_random20,
matched_misleading20. Matched cases independently sample each test's mode with
probability0.5 before observations. Signal accuracy is0.9 except stress cases0.5
and0.1. The first three cases retain v4's fixed mode populations. Matched refers
to the mode prior; the finite plug-in calibration approximation remains.

Two disjoint splits of128 episodes per case. Seed base202609131901 plus fixed
split/case/episode offsets in source; these do not overlap v4. This is1,792
episodes and7,168 scored trajectories, not7,168 independent episodes.

## Predeclared diagnostic outputs

Curve/final Brier, accuracy, confidently-wrong fraction, cost and number of
distinct tests sampled. Paired differences (positive means improvement):

- FF minus IF: change acquisition while keeping fixed scoring;
- FF minus FI: change scoring while keeping fixed acquisition;
- FI minus II: change acquisition while keeping informed scoring;
- IF minus II: change scoring while keeping informed acquisition;
- FF minus II: total change.

Report means and paired z=3.3 descriptive intervals over episodes, split
separately. Do not claim simultaneous coverage, confidence sequences or rare
error bounds from these intervals. A contrast identifies a directional effect
in this diagnostic only when its interval excludes zero. Otherwise inconclusive.
No production-success gate is assigned: this experiment identifies the next
mechanistic lead, not a rescue to adopt. In particular a matched-family pass
cannot override copied-stream harm.

## Verification

Replay every trace and hash; verify each paired scorer's observations agree,
forecast normalization, total cost, no repeated source slots, and compatibility
with frozen v4 diagonal arms in the three retained populations at identical
seeds. Perturbing calibration must not change FF. Existing exact-enumeration
tests cover the unchanged Joint update and unequal-prior extension. Keep all
artifacts exclusive-create and leave prior results untouched.
