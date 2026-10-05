# Frozen entropy-based predictive-information comparison

Freeze before reading coverage efficacy. Use the full decision-time conditional
arrays from the coverage collector, with no additional fitting or label access.
The source is Bickford Smith et al.,
[Prediction-Oriented Bayesian Active Learning](https://proceedings.mlr.press/v206/bickfordsmith23a.html),
AISTATS2023, equations3-5. This is an adaptation to our bounded Bernoulli regime
model and historical unknown-label queries, not a replication of their benchmarks.

For a declared empirical target histogram w, score candidate j by
sum_x w_x [H(p_x)-sum_y m_j(y)*H(p_{x|Y_j=y})], with Bernoulli entropy in nats.
Verify the equivalent weighted Bernoulli KL expression and nonnegative bounds.
Keep the model's original answer masses, rather than the failed shrinkage rule.

Compare three frozen targets: original8 visible inputs153..160, full161 visible
inputs0..160, and recent32 visible inputs129..160. Preserve their matched Brier-
gain controls, random, query entropy, noquery and disjoint8. Tie tolerance1e-10
and smallest origin are unchanged. Every nonempty pool buys one answer. The
actual post-answer publication model and risks remain exactly the same.

Primary outcome remains actual-answer sampled Brier under the established21-cell
phase1 delayed nonharm/transition screen. EPIG targets expected log-score gain,
so a Brier screen is a deliberate downstream test, not its defining identity.
Retain both-answer sampled and population Brier as supplementary diagnostics.
Report all variants and84 cells, no criterion/target tuning or selected winner,
and no untouched, whole-goal, or real-world improvement claim.

Tests: entropy endpoints, zero/positive information, entropy-KL equality, upper
bounds, duplicated-input weighting, answer relabeling, invalid inputs, no source
oracle access. Reconstruct controls and score losses from original artifacts;
hash inputs/code and replay exactly. Time the new scoring component separately
from the costly reference fits, feature generation, retrieval and serving.
Natural-evidence publication changes remain a separate open diagnostic.
