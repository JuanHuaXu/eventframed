# Switching-complexity diagnostic

Consumed v120 forecasts only, all2688 runs/21 cases/2 schedules/2 phases. No
learner tuning, fresh confirmation, policy execution or deployment claim.
Use the same six experts0/1/2/3/10/11 as the prior credit diagnostic and compare
against issued Markov12. Evaluate full256 and terminal64 separately; the latter
allows a new hindsight initial expert and is not continuation of a chosen path.

For expected Brier costs l(t,i), compute exact hindsight minimum with at most
s expert switches, s=0/1/2/4/8. Recurrence:
D(t,s,i)=l(t,i)+min(D(t-1,s,i), min_{j!=i}D(t-1,s-1,j)).
All starting experts are allowed. Also report framewise best discrete expert
and convex-hull projection. Q is used deliberately as diagnostic privileged
information; none of these are valid online forecasts or claimed achievable.

Verify DP against exhaustive path enumeration, monotonicity in switch budget,
best-fixed equality at zero, and hull <= framewise <= budgeted optimum. Hash
inputs/source and require deterministic full replay. Report paired exploratory
mean +/-3.5SE over32 trajectories, not simultaneous/anytime intervals.

Question: does substantial recovery headroom require rapid per-frame switching,
or do few-switch oracle paths already attain it? A positive result is a lead for
an evidence-admissible regime selector, not proof that hidden-Q paths can be
learned from scarce delayed labels. Preserve all stationary and changing cases.
