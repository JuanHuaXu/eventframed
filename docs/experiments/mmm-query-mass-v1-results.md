# Candidate-answer mass diagnostic

## Finding

All2688 records complete. Across1344 delayed records, true population Brier is:

| Selection rule | Brier | Paid queries |
| --- | ---: | ---: |
| No query | .168958621 | 0 |
| Random | .167415352 | 1344 |
| Entropy | .166958523 | 1344 |
| Original joint | .167277354 | 1344 |
| Disjoint joint | .167300670 | 1344 |
| Oracle branch risks, teacher masses | .164737520 | 1344 |
| Oracle branch risks, model masses | .167794066 | 1344 |
| Oracle branch risks, neutral masses | .166295369 | 1344 |
| Oracle risks/teacher masses, abstaining | .164592507 | 1072 |
| Oracle risks/model masses, abstaining | .168464812 | 222 |
| Oracle risks/neutral masses, abstaining | .167289112 | 379 |

Model masses change1073/1344 paid selections relative to teacher masses and
lose .003056545 Brier. The entropy-to-teacher-oracle gap is .002221003, giving
a descriptive ratio1.3762. This is not a causal mediation percentage or a
guaranteed error decomposition. Teacher knowledge is stronger than information
available to any fitted posterior; its advantage is not automatically remediable.

The mean per-pool candidate probability squared error is .045554235. In phase1
switching cases19/20 it is .18010/.17019. Neutral weights improve oracle selection
despite being worse probability estimates in aggregate. Selection depends on
how probability error interacts with differences between answer-branch losses,
not only on overall probability accuracy.

These are NOT online-policy gains. Every branch-risk selector still knows the
teacher and future natural evidence at publication. Abstention rows have different
costs and are not matched-cost comparisons against forced entropy. Two of42
delayed cells have positive descriptive lower bounds for model-mass oracle gains
over entropy; three have negative upper bounds. No whole goal is completed.

## Checks

[Protocol](mmm-query-mass-protocol.md), [raw summary](mmm-query-mass-v1.json),
[exact replay](mmm-query-mass-v1-replay.json) and
[independent audit](mmm-query-weight-v1-audit.json) preserve all original controls.
All11965 distortion identities and5376 paid/abstaining regret bounds pass.
The unit suite checks625 identities/relabelings and3125 regret/order controls,
plus ties, ownership, oracle-input traps and invalid inputs.

The subsequent [single-parameter rescue](mmm-query-shrinkage-v1-results.md)
tests whether this observation yields an implementable query-order improvement.
It does not permit substituting oracle weights in the live forecast law.
