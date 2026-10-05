# Interaction screening feasibility results

13,312 fresh synthetic replicates, 26 cells, completed with byte-identical
replay. Script orthogonality, input immutability and cap assertions passed.
Artifacts: `mmm-spectral-screen-v1.json` and `mmm-spectral-screen-v1-replay.json`.

Target interaction recovery (degree 2 / 3 / 4, each denominator 512):

| Labels | Label noise | Recovered |
|---|---|---|
| 32 | 0% | 512 / 512 / 512 |
| 32 | 10% | 407 / 390 / 387 |
| 32 | 30% | 14 / 11 / 16 |
| 32 | 45% | 0 / 0 / 0 |
| 64 | 0% | 512 / 512 / 512 |
| 64 | 10% | 512 / 512 / 512 |
| 64 | 30% | 86 / 84 / 94 |
| 64 | 45% | 0 / 1 / 0 |

Fair-label controls selected anything in 0/512 (32 labels) and 3/512
(64 labels). In signal cells, any spurious selection ranged from 0 to 7/512.
These are descriptive frequencies, not simultaneous confirmation intervals.

Conclusion: bounded feature enumeration can identify strong parity interactions
without teacher coordinates. It does NOT make weak interactions identifiable
reliably at current budgets: recovery at 30% noise is only 2.1-3.1% with 32
labels and 16.4-18.4% with 64. The threshold is approximately 0.729 and 0.515,
respectively, while that case's population signed correlation is only 0.4.
This predicted power limitation is not a code bug or evidence that no alternative
learner could work. Do not lower the threshold on this outcome and relabel a
rerun confirmation.

Next distinct comparison: a screened interaction logistic learner versus a
regularized all-interaction alternative (or soft shrinkage) with the same label
budget and all original main effects retained. Freeze its full-case quality and
cost protocol before execution. Strong-parity component recovery is not forecast
calibration, shift recovery, selective-evidence validity, or success on any whole
research direction. No production or whitepaper changes. No benchmark performed.
