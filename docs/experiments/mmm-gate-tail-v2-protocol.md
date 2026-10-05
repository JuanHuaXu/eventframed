# Anti-Pigeon gate tail reliability v2

Frozen before running. Keep v1 null, rates, eight starts, equivalence margin,
wealth threshold800, ten scenarios,512 steps and512 streams/scenario/split.
Fresh bases2026103201 and2026103202; same scenario*1000+stream suffix. Existing
v1 files and hashes remain unchanged.

Four gates: fixed0.25; v1 equal-grid; anchored90=0.9*fixed+0.1*grid wealth;
anchored50=0.5*fixed+0.5*grid wealth. These are WEALTH mixtures, not mixtures of
betting rates or p-values. Mixture coefficients are fixed before outcomes.
Each signed wealth is a nonnegative supermartingale under the same conditional
null. Average signs, union over8 starts at800: each arm separately retains1%
anytime bound. Comparing four arms does not give selected-arm familywise1%.

Anchored90 wealth is at least0.9 times fixed wealth, so it crosses no later than
fixed wealth crossing800/0.9. This is not a finite-time delay or miss guarantee:
fixed wealth can cross800 then fall, never reaching800/0.9.

Retain ALL v1 screening conditions, including >=10% improvement in restricted
delay in both moderate shifts and no >10-step harm in other alternatives.
Add tail conditions on EVERY alternative: missed proportion increase <=.01,
and paired normal z3.3 upper bound on miss difference <=.02. Miss includes
premature alerts as in v1. Report missed/premature counts and bounds even when
the mean-delay criterion fails. Null Wilson upper<=.02 remains unchanged.

No tune/selection between splits. Preserve outcomes, first-alert times, hashes,
full replay. These are gate-level criteria only; passing still requires MMM
sharing integration and an independent operational validation before adoption.
