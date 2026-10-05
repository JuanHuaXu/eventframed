# Joint regime-query conditioning component

Frozen before component output. This is a research-only mathematical bridge,
not a quality rescue, allocation policy or serving integration. Earlier
predictive-query v120 used a fixed expert tape; acquisition-training experiments
refitted experts but did not compute joint regime-conditioned query value.
No equivalent joint-conditioning helper was found in the current local research
code. Reuse the existing full segment posterior, not the failed one-change prior.

## Evidence and joint law

At clock t, retain a declared S of at most63 arrived labels, strictly older than
t. Keep every origin/input in the bounded history, but erase unavailable outcomes
from the model view. Retained labels alone carry likelihood factors. Use the
existing per-frame hazard.01, generic family prior.95 and Jeffreys cells. The
same boundary prior and same S are held fixed for all candidate outcomes.

An unobserved query origin j<t is a real past outcome, not a virtual draw from
the current expert. Its two hypothetical labels form S_y=S union {(j,y)}.
There is room for all64 labels; no S member may be evicted. With the same fixed
history and model, m_y=Z(S_y)/Z(S) is the label's conditional probability.
Require m_0+m_1=1 and sum_y m_y Q_y=Q_0 before calling this conditioning.
Reject invalid/known origins or excess support; never repair incoherence by
normalization, clipping or silently dropping old evidence.

For up to eight already visible input probes x, propagate the fitted current
forecast once through the hazard to target t+1: p(x)=.01/2+.99*Q(x). The query
will become available at t+1. Other intervening arrivals are marginalized, not
supplied as known evidence to acquisition. These virtual next-time targets use
visible input values as a proxy, not knowledge of the actual future input law.

Define V(j)=mean_x sum_y m_y (p_y(x)-p_0(x))^2. Independently check equality to
mean_x [p_0(x)(1-p_0(x))-sum_y m_y p_y(x)(1-p_y(x))]. This is expected Brier
Bayes-risk reduction under the declared joint model and probe distribution,
not a guarantee of real-world gain. Changes in other natural feedback and
misspecification can make the practical value smaller or negative.

No actual hidden Y/Q, teacher identity, true boundary or future inputs may enter
these calculations. No old forecast is retroactively revised. Unknown-outcome
poisoning must leave the entire model-visible state and query values unchanged.

## Tests and cost

Enumerate all cut patterns on tiny histories using direct Beta integrals and
check query masses/conditional laws against that reference. Check total
probability, total expectation, the two risk formulas, input ownership, invalid
origins/caps/probes, cancellation, shared-read concurrent equality, and a full
63-plus-one capacity fixture. Race tests precede any quality collection.

Record isolated full-cap base fitting, two-refit query value, and eight-query
pool cost separately. Reuse the existing batched interval builder; no algorithmic
speedup is claimed. For K queries, the initial prototype needs1+2K fits, each
bounded quadratic in label count plus the bounded time-history recurrence. All
of this belongs in the slow path; do not hide hypothetical refits in a constant
or equate microbenchmarks with p95/p99 under serving load.

This component does not complete any of the seven research goals. A later frozen
whole-stream experiment must compare actual learned forecasts at equal outcome
cost against random and entropy sampling, include delays and all stationary/
changed controls, and charge these extra fits. No production, dependencies,
OpenClaw, paper, commit or push changes.
