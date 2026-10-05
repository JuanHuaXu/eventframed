# Fixed-support answer-weight diagnostic

Use every consumed record in publication-v1 and coverage-v1, without new fits,
training, thresholds, or data exclusions. Freeze the ten existing query choices.
Add six diagnostic selectors: minimum population branch risk under teacher,
model, or neutral (.5) answer weights, each forced-paid or permitting abstention.
Use the tested minimumMassRisk implementation with exact minimum and lower-origin
tie breaking. Empty candidate pools select no query. Count every paid action.

Run separately against A's frozen63 and B's frozen64 no-query baseline. Paid
supports and losses coincide; verify forced-paid choices coincide too. B is the
relevant full-budget baseline. Neither state admits newly arriving labels.
Use true teacher-weighted population risk to evaluate all selected actions;
also retain actual-answer population risk and sampled-input risk. The teacher
and branch-risk oracle remain evaluator-only, not available to a real policy.

Check the exact identity R_p - R_q = (p-q)(L_1-L_0), nonnegative oracle regret,
and regret <= |distortion(selected_p)|+|distortion(selected_q)| for paid and
abstaining actions in each state. Reproduce all ten original policies against
the prior publication summary. Preserve all84 cells and both phases separately.
Report how often model and teacher choices differ, probability mean-square error
with equal weight per nonempty pool, costs, risk means, and descriptive paired
mean +/- 3.5 SE contrasts. These are not prospective confidence guarantees.

Interpretation: an oracle branch-risk selector failing after model-weight
substitution shows that response-risk knowledge alone is insufficient for that
selector. It does not prove probabilities are the sole cause of heuristic failure
or that a probability-only rescue exists. No production or whitepaper promotion.
Require exact replay, independent aggregation/identity checks, and source hashes.
