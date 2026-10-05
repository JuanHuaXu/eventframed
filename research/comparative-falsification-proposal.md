# Comparative forecast falsification after v99

Proposal only. v99 passes104/106 quality gates and both finite null checks,
but both parity-to-majority positive recovery gates remain failed. Do not
reclassify near-zero gain as success or tune the6400 boundary on those data.

The neutral alternative detects a model losing to ignorance. It can miss a
model that is worse than another available forecast while still beating0.5.
The consumed v98 diagnosis and v99's remaining gap motivate testing a richer
alternative, not assuming it fixes the gap.

## Predictable alternatives, not independent votes

For each target expert i, keep a neutral alternative with mass1/2 and the
other three raw expert forecasts with mass1/6 each. For each alternative j
and possible start s, accumulate

    M_i,j,s,t = product_{u=s..t} Q_j,u(Y_u) / P_i,u(Y_u).

All Q and P forecasts must exist before the same outcome. Alternatives may
depend on prior data and on the currently observed input; they cannot use the
current or future outcome. Rejected alternatives still count as declared bets,
not as certified truths. Correlation between experts supplies no extra evidence.

Under the target expert's conditional-law null, each factor has conditional
expectation1 because Q_j is a normalized predictable distribution. Mix over
alternatives with the fixed masses above and over32 starts as in v99. The
result remains a nonnegative test martingale. Keep the64-test allocation and
the6400 threshold; do not take the maximum across alternatives without charging
for it. This is a direct specialization of the test-martingale framework in the
[v99 proposal](forecast-falsification-proposal.md), not a new theorem of truth.

The1/2 neutral mass incurs a log2 evidence penalty relative to neutral-only;
other alternatives pay their own mixture penalty. Better alternatives can
increase power, but dilution can also delay a useful rejection. That tradeoff
must be measured, not hidden by reporting only the improved direction.

## Next experiment

First verify a literal alternative/start enumeration, neutral-only reduction,
identical-expert behavior, conditional one-step expectation, extreme full-support
probabilities and the complete publication/feedback lifecycle. Repeat independent
and shared null simulations with fresh seeds. Keep raw-bank, neutral-only and
comparative gates as paired controls on fresh learned streams, retaining all106
quality gates and the same models, priors, fit cadence and evidence budget.

Record when each version is rejected and separate removal of a harmful model
from actual promotion of a useful one. Continue checking all unserved raw
forecasts against the same received outcomes. Do not inherit the raw bank's
cumulative-loss certificate for either gated output. No production promotion,
delayed-feedback or external Anti-Pigeon certificate follows from a component
or finite synthetic pass.
