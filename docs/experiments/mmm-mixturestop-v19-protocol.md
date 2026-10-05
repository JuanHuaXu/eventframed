# Mixture-aware stopping v19

Frozen before evaluation. Three modes: current, full_budget, mixture_stop.
Only adaptive arm2 differs. Mixture_stop honors a guide-confidence stop only
when the actual emitted mixture p is also <=.1 or>=.9. Calculate p from acquired
mask/values and frozen pre-outcome models/weights, exactly as final emission.
No oracle, future label, unobserved input, new data or parameter changes.
Other view selection, confidence thresholds and six-coordinate cap remain fixed.

Use v17 two target families, three input generators, four scenarios, six fits,
two streams per fit, two splits, three modes=1728 streams. Fresh fit2026111201,
design2026111202, confirmation2026111203; existing seed encoding. No tuning.

Keep v17 quality requirements for mixture_stop: every full/post mean harm<=.01
versus current and fixed; clustered majority shift128 post gain>=.005 versus
current in both splits; every shift128 post gain>=.005 versus fixed. Additionally
require mean coordinates<=4 in clustered majority post groups to establish a
cheaper rescue than full_budget. Report full_budget as control, not a fallback
chosen after seeing labels. Preserve all failures and descriptive six-fit95%
bootstrap intervals (20000 resamples, seed2026111299), not simultaneous bounds.

Tests: always-accept/deny gate equivalence to current/full controls, gate errors
and epoch changes reject, final emitted mixture agrees with stop decision,
unchanged arms0/1/3, paired delayed outcomes, complete replay and source hashes.
Gate accesses only acquired coordinates. A confidence agreement is not a truth
certificate; v18 already found confident shared mistakes.
