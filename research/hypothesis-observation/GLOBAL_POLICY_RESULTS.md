# Global policy optimization finds a frozen-forecast barrier

[Protocol](GLOBAL_POLICY_PROTOCOL.md), [three-control game](global-policy-game.json),
[verification](global-policy-verification.json),
[oracle tests](test-global-policy.mjs).

Unlike the earlier shields, this experiment imposes no per-state dominance
requirement. It optimizes mixtures of complete evidence-adaptive acquisition
policies under population constraints across80 fixed source regimes. Forecasts,
observation options, six-query budget, source family and controls remain frozen.

## Numerical bounds

| Screen | Best lower bound | Final mixture upper bound | Witness round |
| --- | ---: | ---: | ---: |
| Random + both entropy versions | .003508657672 | .004135043839 | 107 |
| Random + archived entropy | .002494450436 | .003327302262 | 101 |
| Random + normalized entropy | .002691190575 | .003515346706 | 82 |

The optimized quantity is worst constraint violation: final risk minus control
minus.01, or area risk minus control. A feasible policy would have all final
rows nonpositive and all required area rows strictly negative. Any positive
lower bound on the maximum violation excludes such a policy within this model.

Each lower bound is obtained from one explicit nonnegative row-weight
distribution and its exact-arithmetic linear best-response recursion. It does
not assume the128-round optimizer converged. The remaining primal/dual gaps
therefore do not invalidate a positive lower-bound witness. However, computation
uses floating point, not outward rounding: this is independently reproduced
numerical evidence of infeasibility, not a machine-certified real-arithmetic
proof. The observed numerical discrepancies are many orders below the bounds.

The three-control mixture passes all240 final rows but fails113/225 area rows.
The two-control mixtures fail75 and79 of their310 total rows respectively.
No criterion was loosened and no failing regime was removed.

## Scope isolation

The original protocol used three controls. To check whether the barrier was an
artifact of requiring both entropy implementations, we first projected the
existing witness onto each two-control screen. Those lower bounds were negative
(-.00780317 and-.00406304), hence inconclusive, not evidence of feasibility.
See [preserved projected witnesses](global-policy-reduced-witness.json).

We then reran the same128 rounds and learning rate32 separately with each
entropy version plus random, changing only the control set. Both give positive
lower bounds, independently verified:
[archived entropy results](global-policy-two-entropy.json),
[verification](global-policy-two-entropy-verification.json),
[normalized entropy results](global-policy-two-tie_entropy.json),
[verification](global-policy-two-tie_entropy-verification.json).
These are exploratory scope checks, not fresh confirmation data or a new
predeclared empirical study.

## Why the oracle covers the intended policy class

Under the specified iid-or-root-copy family, count state plus initial reports
is sufficient for future likelihoods and frozen forecasts. Weighted population
loss is additive in unnormalized ordered-history likelihood costs. Backward
recursion minimizes over each observation choice and both possible reports;
probabilities are already included in the state costs. Adding them again would
double-count likelihood. No actual regime is revealed to the policy.

Randomizing among complete policies cannot beat the minimum linear objective
of a deterministic best response. The same weighted lower-bound witness thus
applies to their mixtures, not just the128 policies generated during the game.
It is limited to fixed forecasts, this sufficient-state model and these costs;
it says nothing universal about observation, agency or EventFrame learning.

## Verification

All three games replay byte-for-byte. Each witness was separately recomputed
by recursive state traversal with an alternate squared-loss expression, visiting
48048 states. Independent lower bounds differ by at most1.12e-16. Mixture payoffs
were reconstructed for465/310/310 rows. Forward path scoring and backward oracle
objectives agree within5.56e-16 during optimization. Tiny two-level problems
exhaust64 contingent policies each across32 cost fixtures, plus a tie test.

## Research consequence

Stop searching only for acquisition schedules while freezing this forecast map:
the evidence now indicates a structural incompatibility with the full screen,
not merely an overly conservative local guard. This does not exhaust direction7.

Next lead: optimize forecast and acquisition together. For a fixed row-weight
distribution, the Brier-optimal forecast at a state is the normalized weighted
joint outcome mass. Its minimized risk can enter the same backward observation
recursion. This changes the forecast rather than weakening protection criteria,
and could reveal whether the bottleneck is the fixed forecast map or the
information budget itself. Preserve these frozen-law results as controls.
All seven whole research directions remain open. No production or paper changes.
