# Start-pooled evidence gate v71

Frozen before experiments. Research direction3; not a production Anti-Pigeon
certificate or permission to merge posterior groups. Prior rate mixtures traded
mean delay against deadline reliability. This proposal instead changes how the
eight existing deterministic monitoring starts combine evidence. No rate fitting,
threshold sweep, new outcomes, new starts, resets or source independence claim.

## Contract and proof

Retain v1's conditional null |E[D_t | F_(t-1)]| <= .15, D_t in [-1,1],
predictable evidence pairing, starts at0,64,...,448, signed multiplicative
wealth and fixed rates. Let E_(j,t) be a start's sign-averaged wealth; inactive
starts have wealth1. Each is a nonnegative supermartingale with initial value1.
Their fixed arithmetic mean M_t=(1/8)sum_j E_(j,t) has the same properties,
even though the windows overlap. Alert when M_t>=100; Ville bounds the per-arm
ever-alert probability by.01 under the stated null. Do not average log wealth
arithmetically or multiply dependent evidence; evaluate log of arithmetic mean.

The old any-start rule alerts when max_j E_(j,t)>=800. Such a crossing implies
M_t>=100. Thus the pooled first alarm is no later, pathwise. This does NOT
guarantee fewer scored misses after a change: an earlier false alarm counts as a
premature terminal failure. Preserve that failure and never relabel it detection.

Sources: [Vovk & Wang, E-values: Calibration, combination, and applications
(2021)](https://arxiv.org/abs/1912.06116) for arithmetic evidence merging;
[Howard et al. (2021)](https://arxiv.org/abs/1810.08240) for time-uniform
inference. The finite eight-start implementation/proof above is our adaptation.
The test does not prove the null holds for real paired evidence.

## Frozen screen

Four arms on identical tapes: old fixed, old grid, pooled fixed, pooled grid.
Reuse v1/v2's ten scenario generators,512 steps and512 streams per scenario per
phase. Design base2026117101 and confirmation base2026117102. Seed is base*1e6
+ scenario*1000 + stream. No tuning between phases. 10240 streams total.

Retain v1's rules: each null Wilson95 alert upper<=.02; both moderate shifts
require >=10% restricted-mean improvement over fixed with paired z=3.3 lower
bound>0 and no increase in premature count; other alternatives permit no more
than10 steps restricted-delay harm. Misses/premature alarms cost the entire
remaining horizon. Report detected-only delay separately.

Retain v2's additional miss limits: observed excess miss fraction<=.01 and a
one-sided bound<=.02 in each alternative. Replace its degenerate normal bound:
bound the probability of a harmful discordance (candidate misses, fixed detects)
with a one-sided Clopper-Pearson bound. Excess miss probability cannot exceed
this probability. Use alpha=.05/30 for three candidates x five alternatives x
two phases; frozen independent synthetic trajectories support binomial sampling.
Report design and confirmation, apply adoption screen to confirmation as before.
This more conservative simultaneous bound is not a real-stream certificate.

Both old controls must replay exactly, and pooled first alarms must dominate
their corresponding old first alarm on every tape. Test direct wealth arithmetic,
inactive-start unit mass, invalid-input state preservation, sign symmetry and
threshold behavior. Record full tapes, source/protocol hashes and all counts.
All arms get separate per-arm theoretical1% bounds, not a combined1% budget for
post-hoc choosing/running every arm. Gate-only success needs fresh integrated
sharing validation. Keep serving unchanged regardless of this screen.
