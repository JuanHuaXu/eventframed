# Candidate-answer mass substitution diagnostic

Freeze before evaluation. This is an offline causal-factor isolation, not an
online acquisition policy or a population causal claim. The weak critics could
reflect candidate-answer probability error, inaccurate estimates of conditional
forecast usefulness, or unpredictable future natural evidence. Change only the
answer weights while holding all fitted counterfactual branches fixed.

Use all2688 consumed population-query records and the matching decision-time
query masses from the disjoint-query artifact. For each paid candidate j, let
L_j(0),L_j(1) be its saved population risks, q_j the teacher answer probability,
and p_j the strictly as-of160 model mass. Compare exact risk minimization using
q_j, p_j and the neutral .5, both forced-paid and with no-query abstention.
Evaluate every selection under q_j. Preserve noquery/random/entropy/joint/
disjoint controls. All selectors still see oracle branch risks; substituting p
does not make them operational. No-query and naturally redundant branches have
identical losses for both answers and therefore cannot depend on the weight.

Check the exact identity R_p(j)-R_q(j)=(p_j-q_j)*(L_j(1)-L_j(0)). For a common
action set, let j_p and j_q minimize R_p and R_q. Then true regret satisfies
0<=R_q(j_p)-R_q(j_q)<=|R_p(j_p)-R_q(j_p)|+|R_p(j_q)-R_q(j_q)|.
The proof adds/subtracts the two modeled risks and uses R_p(j_p)<=R_p(j_q).
Tie-break by smallest origin; no-query origin-1 wins exact abstention ties.

Report all84 cells, per-trajectory mean candidate squared probability error,
selection changes, regret and the fraction of the entropy-to-teacher-oracle gap
lost by mass substitution. Ratios of aggregate differences are descriptive and
are not mediation fractions or guaranteed additive error decompositions.
Report mean +/-3.5SE over32 trajectories for paired cell differences, not a
simultaneous/anytime confidence claim. Retain complete-delivery identities.

Tests: exhaustive small branch controls, equal weights, equal answer losses,
answer relabeling, order/tie invariance, input ownership, invalid probabilities,
strict selector input boundary, distortion identity and regret bounds. Audit
decision masses from original sources, reconstruct all earlier population
controls, hash every input and code dependency, and replay byte-exactly.
No model fitting, production/paper changes, tuning or whole-goal promotion.
