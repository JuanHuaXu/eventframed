# Exact renewal-allocation stress test

Frozen after the single-schedule mixed-counterfeit control, before this sweep.
Do not change models, priors, noise20 or total16 credits. Four ordinary first
reports are from types0,1,2,7. Allocate six renewal reports among those four
types, at least one to each. Enumerate all10 possible count vectors. Final
posterior is order invariant under the declared model; within each allocation
use ascending type order and increasing per-type slot numbers.

For each allocation enumerate all1024 binary report vectors and all16
counterfeit masks across the four measured types. The former single schedule's
all-genuine/all-counterfeit/even/odd cases are included. This does not increase
sample size until a desired p-value appears: all cases are exact finite-law
expectations, with no parameter selection or model retuning.

For each of160 allocation/mask cells require hierarchical Brier no worse than
local by more than.01. For each all-genuine allocation retain gain>=.005, and
for each all-counterfeit allocation retain confidently-wrong reduction>=.05
against certain freshness. All180 gates required for this stronger stress screen.
Record failures, worst harms, exact population masses and oracle floors. A
passing or failing result does not rewrite the earlier six-gate test.

Verify all allocation/mask masses, oracle lower bounds, total credits and paid
slot validity, plus source hashes and complete replay. Classifier tie decisions
use the existing lowest-index Python argmax. No production or whitepaper edits.
