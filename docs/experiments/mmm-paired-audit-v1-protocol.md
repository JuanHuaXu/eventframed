# Costed paired audit v1

Compare evidence collection, not acquisition policies. Retain the consumed
logged-gain-v1 realized-outcome projection and both random/entropy and
joint8/entropy contrasts. No generator, future outcome or query strategy changes.

For each contrast, an episode with identical query origins has exactly zero
gain and requires no comparison query. Otherwise:
-Single action: choose one of the two strategies with probability.5, pay one
 query, observe only its loss. Compare IPS and phase0-only constant DR.
-Paired audit: include the episode with probability.5, pay two distinct queries
 and fit/score two isolated branches; otherwise pay zero and observe neither
 loss. Compare HT and phase0-only difference-regression DR. Both branches share
 the same later realized outcome, not each other's acquired training label.

Each design spends one expected query per nonidentical episode. This is expected
cost matching, NOT equal realized budgets; report actual cost ranges and do not
enforce a hidden outcome-dependent quota. Scoring both queried branches also
costs more than scoring one, and outcome availability is assumed by this
offline simulator. Future outcomes are fixed independently of audit actions;
this is not a causal estimate for an agent that changes its environment.

Use64 logging assignments with SHA256 rejection-sampled fair coins keyed by
replicate/phase/case/index/contrast/design. This is simulation randomization,
not64 independent datasets. Use672 phase0 delayed episodes to train regressions
and672 phase1 episodes to evaluate. Single means have prior sum.5,count1;
paired difference has prior sum0,count1. Update regressions only with observed
phase0 outcomes. Exact-identity episodes do not supply fabricated regression
observations. Regressions freeze before phase1.

Paired DR increment is m+J/q*(D-m), q=.5 and D=control loss-candidate loss.
No inclusion means increment m and no observed D. HT sets m=0. Known inclusion
probabilities are required; source/model confidence is not a substitute.
All increments must lie within[-3,3]; reuse the unchanged fixed-scale4 EB CS,
rate grid and two-sided implementation. Allocate alpha=.05/8 to each of the
two contrasts x two collection designs x two estimators. All eight are tested
in one declared family, including the single-action controls rerun at this
coverage allocation.

Report fixed-table gain, estimator RMSE, final width, every-prefix coverage and
positive lower bounds, nominal query costs including regression training and
number of paired fits required. Test exact expectation identities, skipped-loss
access traps, same-origin zero evidence, oracle separation, original point
targets, per-prefix independent audit and deterministic replay. A measurement
win does not rescue negative policy gains or satisfy prior all-case quality
criteria. No production, paper, installs, commits or pushes.
