# Frozen candidate-mass shrinkage rescue

Motivation: the mass-substitution diagnostic finds model weights lose the oracle
opportunity and neutral weights perform better. This is not proof that a learned
calibration map will rescue acquisition. Test one parameter, no feature or
threshold sweep, before expanding the model.

Use p_alpha=.5+alpha*(p-.5), alpha in[0,1]. Fit alpha on phase0 delayed candidate
rows with equal weight per pool by minimizing squared loss:
alpha=clip(sum w*(p-.5)*(y-.5)/sum w*(p-.5)^2,0,1).
If the denominator is zero set alpha=0. Freeze before phase1. The main fit uses
actual revealed labels y, available offline in the synthetic training archive,
not future labels at inference. A separately labeled diagnostic fit substitutes
known generator q for y; it is not validation on new real labels.

Query score is mean_probe[(p1-p0)^2]*p_alpha*(1-p_alpha), retaining the original
as-of160 conditional predictions and probes153..160. At alpha=1 this reproduces
the existing joint-value score. The score is the variance of a declared
two-answer predictive mixture; its implied marginal generally differs from the
original posterior marginal when alpha differs from1. Treat it as an acquisition
heuristic, NOT a certificate of proper-risk improvement for the incumbent law.
The served/updated forecast remains the existing separately fitted query branch.

Arms: noquery, random, entropy, original joint, label-trained shrinkage,
teacher-target-trained shrinkage (diagnostic), neutral .5, and direct teacher q
(oracle diagnostic). All nonempty pools buy exactly one answer, no new abstention
threshold. Evaluate actual-answer sampled Brier, both-answer sampled expected
Brier, and both-answer population expected Brier using all saved branches.

Primary advancement: same consumed phase1 delayed21-cell screen as earlier,
lower gain bound>=-.001 versus random and entropy in every cell, and positive
lower bounds for cases19 and20 versus both. Mean +/-3.5SE across32 trajectories
is descriptive, not simultaneous/anytime inference. Report every arm, including
failures, and all84 cells. Do not choose alpha or a model on phase1. An oracle
arm cannot advance a deployable policy. No whole-goal or untouched-data claim.

Verify analytic coefficient against direct convex objective, degenerate and
boundary cases, alpha1/raw-control identity, selector oracle-field isolation,
phase1-label poisoning, equal costs, input ownership, all loss rescoring, source
hashes and exact replay. Measure selection with precomputed conditional outputs;
exclude Bayesian fitting/features, retrieval, I/O and serving from those timings.
No production, paper, commit or push changes.
