# Frozen interval-aware acquisition pilot

The old900-gate fixed-observation screen remains FAILED and has a proven
unattainable gain cell. Return to goal7's equal-cost acquisition requirement:
change what is observed, not that old screen's thresholds or recorded status.

Three active policies share one inference model: target-class expected Gini
reduction, uniform random, and maximum next-report binary entropy. A fourth
fixed schedule [0,1,2,7,0,1] is diagnostic. All begin with one ordinary report
from types[0,1,2,7] (cost1 each), then purchase six renewals (cost2 each).
All spend16 credits. No extra observations or free reliability signals.

The joint model uses uniform h in0..15, uniform common noise on[.10,.30], and
mask prior .25 all-genuine + .25 all-copied + .50 uniform over sixteen masks.
Integrate joint masses using the existing Bernstein polynomials BEFORE
normalizing. Recompute candidate branches with hypothetical binary outcomes.
Actual hidden h, noise, mask and unobserved tapes are never passed to selection.
Ties use source order. This is one-step target-class observation design, not HRM
training, grokking, unrestricted hypothesis invention or validated provenance.

New seed strings use base2026091401, two splits, three noise levels(.10,.20,.30),
four masks(0,5,10,15),64 episodes/cell/split:1536 episodes. Presample a common
root/fresh uniform tape for all policies. Store complete paid-query traces,
forecasts, actual outcomes and seeds. Separate actual masks and latent h are
used by the generator/scorer only. No seed search or post-result prior tuning.

Record final Brier, accuracy, confidently-wrong fraction(max probability>=.9),
and post-initial learning-area Brier: sum of each pre-renewal forecast loss
times2/12 across the remaining12 credits. Initial4 costs are common and excluded
from this area statistic, but included in total cost.

In EACH noise/mask/split cell require candidate final-Brier nonharm against
random AND entropy: paired lower gain>=-.01 (48 gates).
In each genuine/mixed mask(0,5,10), require post-initial area gain lower>0
against BOTH controls (36 gates). All84 must pass. No positive-gain requirement
on fully copied renewals: those supply no new independent draws.
Intervals are paired mean +/-3.3 standard errors over64 episodes, descriptive
normal approximations, NOT simultaneous coverage or confidence sequences.
Report failure rather than giving that approximation stronger guarantees.

Verify joint integration against direct quadrature, action/outcome separation,
prefix replay, equal costs and original trace reproducibility. Full rerun must
be deterministic. Finite success would not complete real-task or production
requirements. No production access, private data, whitepaper promotion or push.

