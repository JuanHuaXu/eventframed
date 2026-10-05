# Regime-safe observation policy improvement

Exploratory design on the consumed grid. Do not call it fresh validation.
The predictive law stays fixed. Build one state-dependent policy without access
to the realized noise or mask. Its declared family contains five noise levels
(.10,.15,.20,.25,.30) times all16 source-copy masks. The actual fixed regime
remains unknown to the policy, and impossible histories under a regime are
excluded only where their likelihood is exactly zero.

Reference policy: tie_entropy. For each supported regime g and count state s,
compute reference final loss F_g(s) and remaining pre-query loss sum A_g(s).
At terminal states F is the expected Brier loss of the frozen forecast and A=0.
Work backward, maintaining corresponding candidate values f_g,a_g.
For each action u, compute candidate-continuation final loss
f_g(s,u)=sum_y p_g(y|s,u) f_g(s+u,y), and pre-query loss
a_g(s,u)=loss_g(s)+sum_y p_g(y|s,u) a_g(s+u,y).

An action is eligible only if f_g(s,u)<=F_g(s)+r*1e-12 and
a_g(s,u)<=A_g(s)+r*1e-12 for EVERY supported regime, where r is remaining
queries. The tolerance is numerical slack, not the experiment's .01 allowance.
Among eligible actions minimize the original model-prior post-query Brier sum
with candidate continuation and the same 1e-12 lowest-index tie rule.

In exact arithmetic the reference action is feasible by backward induction:
candidate continuations no worse than reference in each regime imply that
following its current action preserves both inequalities. With the declared
slack the child allowance is (r-1)*1e-12, leaving one increment for arithmetic.
Assert feasibility rather than silently loosening tolerance on a failure.
This is a sufficient policy-improvement construction, NOT global minimax
optimality. The induction protects only this finite family, not off-grid noise.

Evaluate exact population final and area Brier in all80 worlds. Preserve .01
final nonharm and strictly positive area gains against random and both entropy
controls (area excludes all-copy mask15). Report zero/roundoff gains separately;
they are not evidence of useful improvement. No weak-dominance relabeling of
the original success criteria. No empirical coverage or source authentication.

Motivation: robust dynamic programming studies transition-model ambiguity;
see [Iyengar (2005), publisher abstract](https://pubsonline.informs.org/doi/pdf/10.1287/moor.1040.0129).
Its rectangularity-based optimality theorem is not invoked here. Our uncertainty
is a fixed global source regime, and the finite induction above is the claimed
argument. Publisher metadata/abstract was available; full-text retrieval of the
related Nilim/El Ghaoui paper returned403, so no full-paper fidelity is claimed.

Verification: replay, unchanged-control parity, separate forward versus backward
scoring, all supported state/regime feasibility checks, and branch normalization.
Runtime code and archived sources remain unchanged. A promising finite result
must be followed by off-grid and fresh-task evaluation.
