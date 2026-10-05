# Mixture-aware acquisition v7: FAILED rescue screen

Completed1,792 fresh episodes and7,168 scored trajectories under the frozen
`MIXTURE_ACQUISITION_PROTOCOL.md`. Raw data/source hashes are in
`mixture-acquisition-v7.json.gz`; all results are in the corresponding summary
JSON. Every arm receives16 reports plus eight signal checks, cost24. No serving
implementation or performance threshold was changed.

## Confirmation outcomes

The acquisition-speed gate failed. Independent20 curve Brier was0.341851 for
mixture-aware acquisition versus0.341226 for fixed acquisition with the same
averaged scorer: gain-0.000625, paired z3.3 interval[-0.009160,0.007911]. The
required gain was>=0.005 with a positive lower bound. Final Brier was likewise
nearly unchanged,0.141309 versus0.141129. The first split's curve gain was
-0.002760 with an interval spanning zero.

The misleading-signal rescue gate also failed its uncertainty condition. Final
Brier was0.317851 versus0.367909 reliable-only: mean gain0.050058 passed0.03,
but interval[-0.011282,0.111399] crossed zero. Compared with fixed acquisition
and averaging, gain was-0.001673 with interval[-0.019996,0.016650]. There is no
evidence here that the new selector explains the rescue retained in the mean.

Mean curve/final harm ceilings against fixed passed in all seven confirmation
cases. Independent20 mean harm versus reliable-only was0.005968 curve and
0.006169 final, also below0.01. These finite mean checks do not establish
confidence-certified non-harm or negate the two failed benefit gates.

| Case | Fixed acquisition + averaging final Brier | Mixture acquisition + averaging final Brier |
| --- | ---: | ---: |
| Independent20 | 0.141129 | 0.141309 |
| Copied20 | 0.415913 | 0.415026 |
| Mixed20 | 0.213296 | 0.206926 |
| Matched05 | 0.037734 | 0.037595 |
| Matched20 | 0.298984 | 0.298707 |
| Matched random signal20 | 0.410172 | 0.407967 |
| Matched misleading signal20 | 0.316178 | 0.317851 |

## Verification

`python3 -m unittest test_mixture_acquisition_v7 test_reliability_v6 -v`
passed all five tests. The new analytical acquisition gains agree with explicit
copied-state updates for both counterfactual outcomes at all eight tests after
a history with repeated and conflicting reports. Gain nonnegativity (within
floating-point tolerance), exhausted slots and no mutation during selection pass.
Full replay of both datasets, hashes, environmental tape agreement across arms,
unique source slots, normalized forecasts and identical budgets pass. The v6
joint-enumeration tests still pass.

## Interpretation and next discriminating check

This tested the actual joint-mixture one-step target objective, not independent
averaging of source priors or a source-entropy heuristic. It did not resolve the
learning-speed tradeoff at the same budget. Keep the fixed selector as control;
do not enable the more expensive selector merely because its equation is coherent.

Before building another large rescue, inspect bounded two-step action values on
recorded pre-outcome states: does allowing a second observation change the
preferred action enough to justify a fresh equal-budget rollout? Model-implied
lookahead value is a diagnostic only, not empirical gain. If values/actions are
almost unchanged, move to cost-sensitive stopping or a different observation
family instead of increasing lookahead blindly. A larger evidence budget must
remain a separately labeled information diagnostic, not a replacement success gate.
