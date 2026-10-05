# Frozen population-risk allocation experiment

Replace per-history conditional protection with the original population harm
limit .01. Keep the fixed-.20 local baseline and uniform-noise optimistic
target from the preceding candidate. Each of1024 report histories gets one
lambda in[0,1], specifying p_x=p0_x+lambda_x(target_x-p0_x).

For each copied-root mask, form the population regret polynomial in common
noise on[.10,.30]. Elevate all likelihood coefficients to degree10 before
summing over histories. Constrain every Bernstein coefficient of that
population regret to <=.01. This suffices for every noise in the interval,
but can be conservative. It does NOT bound each individual forecast's harm.

For each mask/coefficient j:
  R_j(lambda)=sum_x[A_jx lambda_x^2+B_jx lambda_x] <=.01.
A_jx is joint history mass coefficient times squared correction norm;
B_jx is twice the correction dotted with (mass*p0 - joint class masses).
A_jx>=0. Minimize uniform-noise all-genuine population regret, obtained by
averaging the eleven mask0 coefficient rows.

Fit each allocation against the full declared law family, not the actual
evaluation world. Freeze the resulting history-to-lambda mapping before
scoring the five actual-noise worlds and sixteen masks. Runtime receives only
the observed history and known schedule. No actual latent hypothesis, true
noise or true mask enters that mapping. This is exact finite-model design on
consumed cases, NOT fresh confirmation or a scalable learned memory adapter.

Preserve all900 gates:800 harm<=.01,50 genuine gain>=.005,50 false-confidence
reduction>=.05. Source controls stay identical. No threshold tuning.

## Solver contract
Use nonnegative Lagrange multipliers with exact box-constrained separable
quadratic minimizers and coordinate dual maximization by bisection. The
all-zero lambda is strictly feasible. Repair any residual numerical primal
violation by a uniform contraction toward zero, then compute the repaired
objective minus the valid dual lower bound. Require gap<=1e-8 and maximum
constraint<=.01+1e-12. Stop after at most500 sweeps; a failed solve terminates
the experiment rather than emitting unqualified results. No closest-point
solver or tolerance change to older experiments.

Verify analytic small problems, feasible-grid comparisons, likelihood mass
normalization and coefficient/direct-risk agreement. Replay the experiment,
retain original controls and independently recompute squared losses.
A finite pass does not establish real-data model coverage or closed-loop
policy benefit. No production changes or publication.

