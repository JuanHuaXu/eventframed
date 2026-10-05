# Remaining-horizon budget: partial acquisition improvement

[Protocol](RISK_BUDGET_PROTOCOL.md), [evaluation](risk-budget-evaluation.json),
[verification](risk-budget-verification.json).

This consumed-grid diagnostic changes only the final-loss eligibility budget.
The forecast, prior objective, source family, controls and tie rule stay fixed.
The budget is .01 over six queries, allocated uniformly by remaining horizon.
It is a sufficient conditional shield, not optimal trajectory-budget allocation.

## Results

All240 final nonharm checks pass. Positive-area checks improve from116/225 for
the zero-budget shield to156/225. The full screen still FAILS69 area conditions.

| Control | Final nonharm /80 | Positive area /75 | Worst final gain | Best final gain |
| --- | ---: | ---: | ---: | ---: |
| Random | 80 | 70 | -0.0018194411 | 0.1360396307 |
| Archived entropy | 80 | 46 | -0.0054374982 | 0.0069231625 |
| Normalized entropy | 80 | 40 | 0 | 0.0021546984 |

Against normalized entropy,40 area gains exceed1e-12 and35 are exactly zero;
none are negative. Maximum area gain is .0003417352. These absolute Brier-unit
gains are small and must not be described as comparable accuracy percentages.
Against random, five area harms remain; against archived entropy,29 remain.
The original .01 final allowance and strictly positive area criteria are intact.

For all-genuine sources, the five area gains are .0000174358, .0000266228,
.0000411877, .0000600349, .0000806077 at noise .10/.15/.20/.25/.30. All five
are positive, but this is an exact finite-model design result, not empirical
confirmation of actual-agent benefit.

The new policy changes1664 nonterminal states relative to normalized entropy,
versus1440 for zero budget. Counts by prior renewals are
[0,0,0,96,208,1360]. Improved area proves that some effective changes now occur
before the last observation. These counts include unreachable states and are
not frequencies. The largest conditional final-risk excess in any supported
state is .0016107492; this is not the population excess. Population final gains
against normalized entropy are nonnegative in every tested world.

## Verification and interpretation

Zero budget exactly reproduces all20592 archived shield actions and its full
compiler statistics. Three invalid-budget tests pass. Full replay matches byte
for byte,880 old-policy/world score objects remain identical,360 separate
backward score checks agree within2.98e-14, and16896 direct likelihood checks
agree within1.39e-17. The compiler checks668880 supported state/regime budgets
and2675520 branch sums; forward evaluation checks6720 probability masses.

This supports the hypothesis that zero conditional risk tolerance was suppressing
useful observations. It does not establish that uniform risk allocation suffices:
the learning-speed success criteria still fail. Source-family adequacy and
off-grid protection are not proven by the on-grid induction.

Next: freeze this candidate and test intermediate noise values not used by its
guard before further tuning. More expressive allocation across branches remains
a separate lead. No production, whitepaper or default behavior changed; all
seven research directions remain open.
