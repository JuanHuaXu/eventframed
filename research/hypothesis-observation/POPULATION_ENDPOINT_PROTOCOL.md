# Frozen upper-noise-endpoint objective

The uniform-noise average objective passed888/900 gates but left12 genuine-gain
misses at .25/.30 noise. Change only the design objective to all-genuine Brier
regret at the fixed upper endpoint .30. The degree10 Bernstein coefficient
k=10 of mask0 is exactly that endpoint population regret.

Keep the same uniform-noise optimistic target, local baseline, lambda box,
all176 population coefficient constraints (epsilon=.01), solver5000-sweep cap,
1e-8 dual gap and all900 quality gates. No true evaluation noise or mask enters
the fitted mapping; .30 is a fixed design choice for ALL evaluation worlds.
This is a consumed-case rescue motivated by exposed failures, not fresh
confirmation, minimax optimization or proof that .30 is worst for every rule.

The pointwise-harm tradeoff of population protection remains. Record it.
Replay all optimization, compare original controls, and verify direct versus
polynomial risk and alternate Brier scoring. If endpoint optimization damages
other regimes or still fails the gain floor, report every failure.

If a solved allocation cannot gain .005 at .30, its checked primal-dual bound
also supplies a numerical upper bound on any gain attainable at .30 by this
SAME line-restricted family and coefficient constraints. Do not generalize that
ceiling to different correction directions, observations, or all feasible
population-risk policies. No numerical tolerance changes or production work.

