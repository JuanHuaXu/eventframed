# Sample-count-matched regime diagnostic

Frozen before results. This is a hindsight diagnosis, NOT an online policy.
Use cases0 (stationary),19 and20 (opposite Boolean transitions) from consumed
v120, both phases/schedules,32 trajectories each. Snapshots160,192,224.
The known128 boundary is a real change in19/20 and a placebo cutoff in0.
No detector is supplied with this boundary and no research goal is completed.

At each snapshot, let F be the original latest64 naturally arrived labels.
Let C be the subset of F with origin>=128, with n=|C|. Refit the frozen.95
generic-prior family on F, C, and eight uniform-pseudorandom subsets of F of
size n. Controls use SHA256 priorities keyed by the literal
family-regime-match-v1:phase:case:index:clock:replicate:origin, ascending hash,
origin tie-break, then fit in ascending-origin order. Omit schedule from the
hash for common priorities across paired schedules. Labels/Q never select
random subsets. An empty subset has the prior predictive0.5, not a fake label.

Score each fixed snapshot on its next32 actual query inputs with hidden Q used
ONLY for Brier evaluation. These correspond exactly to the original32-cadence
publication windows. Retain all ten predictions, origins, likelihoods and family
weights. Refit F must match the prior family artifact, not merely its mean score.

Primary diagnostic contrast: mean random-control risk minus C risk, averaging
the eight random draws WITHIN each trajectory before uncertainty over32 distinct
trajectories. Also report F-minus-C and F-minus-random. This controls sample
count, not every possible covariate/recency difference. Pair with stationary
placebo results; do not automatically attribute improvements to causal regime
purity. Mean +/-3.5SE is exploratory, not simultaneous/anytime coverage.

Verify exact counts, origin eligibility, selection reproducibility, no repeated
origins, empty-prior behavior, all-current equality, future-label/Q poisoning,
full-model reference parity and independent scoring/replay. Use1152 snapshots
(384 trajectories x3), not9216 independent random-control observations.
No prior tuning, changed success gates, production changes or learned boundary.
