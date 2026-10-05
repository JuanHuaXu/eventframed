# Paired excess-risk bound v73

Statistical-method validation and explicitly post-hoc sensitivity analysis of
v72. No detector, acquisition policy, seed, original pass label or frozen v72
criterion changes. This cannot count as a fresh confirmation experiment.

For iid paired trajectory outcomes define H=1 when candidate misses and control
detects, B=1 for the reverse; they are mutually exclusive. Let s=P(H or B),
r=P(H | H or B). The excess miss probability is delta=s(2r-1). Given n pairs,
d=h+b ~ Binomial(n,s), and h | d ~ Binomial(d,r). Build three one-sided
Clopper-Pearson bounds, each with error alpha/3: lower/upper bounds s_L,s_U
for s and upper bound r_U for r. For d=0 set r_U=1. Then

U = s_U(2r_U-1) if r_U>=1/2, else s_L(2r_U-1)

is an upper confidence bound for delta with noncoverage at most alpha. On the
intersection of those three coverage events, the expression is maximized by
r_U and by the appropriate endpoint of s. A union bound needs no independence
between the three bounds. Conditional coverage for r given each d implies
unconditional coverage. The source building block is
[Clopper & Pearson (1934)](https://doi.org/10.1093/biomet/26.4.404).
This particular conservative rectangular construction is our derivation, not a
claim of optimality or a method attributed verbatim to that paper.

It accounts for both harmful and beneficial discordances. It need not dominate
the old harmful-only bound in every sample. It is not a confidence sequence:
sample size must be fixed in advance, iid trajectory sampling must be justified,
and adaptive model/data selection needs a separate analysis. Within-trajectory
dependence is permitted; arbitrary dependence between trajectories is not.

Validate analytic binomial special cases, input limits, interval ordering,
zero-discordance nonzero uncertainty, and enumerate the full multinomial sample
space for n=1,4,8,16,32 at alpha=.05,.01 on all admissible pairs from probability
grid {0,.01,.05,.1,.2,.4,.6,.8,.95,1}. Verify total mass and noncoverage<=alpha
numerically. This finite grid checks implementation, not the general proof.

After implementation validation, analyze the twelve corrected-arm alternative
cells of the consumed v72 artifact with alpha=.05/12. Record both bounds and
the original decision, plus raw artifact and analysis/test/protocol hashes.
The familywise interpretation assumes a predeclared application; because this
analysis was motivated by v72, report it as diagnostic and require fresh data
before operational claims. Do not replace v72's failed10% speed criterion.
