# Regime-safe acquisition: protection without faster learning

[Protocol and finite induction](REGIME_SAFE_PROTOCOL.md),
[population results](regime-safe-evaluation.json),
[verification](regime-safe-verification.json),
[decision-depth diagnostic](regime-safe-depth-diagnostic.json).

This is a consumed-grid design experiment, not independent confirmation or a
production implementation. It retains all forecasts and changes only acquisition.
The single policy has no access to the realized source mask or noise level.

## Result

All 240 final-loss nonharm comparisons pass against random, archived entropy,
and tie-normalized entropy. The original allowance stays 0.01. However, only
116/225 positive learning-area conditions pass: the full research screen FAILS.

| Control | Final nonharm /80 | Positive area /75 | Worst final gain | Best final gain |
| --- | ---: | ---: | ---: | ---: |
| Random | 80 | 70 | -0.0018194411 | 0.1349657416 |
| Archived entropy | 80 | 46 | -0.0054374982 | 0.0058806135 |
| Tie-normalized entropy | 80 | 0 | 0 | 0.0007912929 |

Against normalized entropy, all area gains are exactly zero in the saved output,
not small positives rounded away. Against random there are five genuine area
harms; against archived entropy there are29. None are converted into successes.
The largest area harms are .0052664650 and .0068220973, respectively.

On all-genuine sources, final Brier gains over normalized entropy range from
.0000328763 at noise.10 to .0002501714 at noise.30. These are modest absolute
score improvements, not percentages of accuracy or evidence of general learning.

## Why the gain is limited

The guard requires conditional final loss and remaining learning-area loss to
be no worse than normalized entropy in every supported declared regime at every
state. It is stronger than the original population nonharm requirement.

Of20592 nonterminal count states,1440 change action. Counts by number of prior
renewals are [0,0,0,64,192,1184]. Restricting to states reachable under the
candidate gives [0,0,0,0,0,24]. Thus every reached action change is at the sixth,
last query. It can improve terminal forecasts but cannot improve the six
pre-query losses defining learning area. Reachable-state counts are not visit
frequencies or sample sizes.

This diagnoses conservatism in the sufficient guard. It does not prove that
faster protected acquisition is impossible, nor that globally constrained
policies must obey this stronger statewise condition.

## Checks

- 668880 supported state/regime checks: maximum final and area excess both0.
- 2675520 conditional branch-normalization checks in policy construction.
- 6160 world/policy/stage mass checks in forward evaluation.
- Byte-exact replay and800 bit-identical archived-control/world score objects.
- 16896 direct likelihood checks, maximum discrepancy1.39e-17.
- 330 separate backward-score checks, maximum discrepancy2.98e-14.
- Four inherited tie-rule tests, including invalid input.

The backward verifier shares the saved model family and compiled policy with
the forward evaluator; it independently computes propagation and squared loss,
not the correctness of the assumed family. No certificate outside the five
noise points or these sixteen fixed source masks is claimed.

## Next lead

Test a trajectory-level risk budget that permits a temporary conditional
disadvantage while maintaining the original population nonharm allowance and
positive learning-area criterion. Do not simply relax the success thresholds.
Preserve this zero-budget guard as a control. Off-grid evaluation and prospective
agent evidence remain required before any promotion. All seven directions stay
open; no runtime, production or whitepaper changes were made.
