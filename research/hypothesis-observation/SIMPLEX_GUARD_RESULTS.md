# Outcome-simplex guard: protection rescued, useful gain not rescued

The [frozen test](SIMPLEX_GUARD_PROTOCOL.md) passes 850/900 gates, hence FAILS
the combined screen. All 800 nonharm and all 50 false-confidence gates pass;
all 50 genuine-gain gates fail. There is no threshold change or discarded case.

| Actual noise | Pass /180 | Worst population Brier harm | Genuine Brier gain range |
| --- | ---: | ---: | ---: |
| .10 | 170 | .000158 | .001849 to .002039 |
| .15 | 170 | .001034 | .001617 to .001903 |
| .20 | 170 | .001866 | .001210 to .001547 |
| .25 | 170 | .002661 | .000591 to .001169 |
| .30 | 170 | .003424 | -.000178 to .000668 |

The .01 harm ceiling is now protected without the measurement likelihood being
correct, within the fixed four-outcome scoring space. Useful improvement is not:
all gains miss .005, and some .30 genuine cells get worse than local inference.

## Why a different target cannot rescue this guard in the matched world

For history x, let q(x) be the true conditional four-class law and F(x) the
simplex forecasts whose Brier regret versus local p0(x) is <=.01 for each of
the four individual outcomes. Conditional expected Brier risk equals

    R(p;q) = ||p-q||^2 + 1 - ||q||^2.

In the matched .20, all-genuine world, the optimistic proposal already equals
q(x), checked by the runner. Its projection onto F(x) therefore minimizes
conditional risk over ALL forecasts satisfying these per-outcome constraints.
This remains true for randomized feasible forecasts: averaging cannot improve
on the minimum conditional risk. Summing over histories gives the population
optimum for this fixed baseline, observation schedule and feasible family.

The numerical projection objective uses half squared distance. Its maximum
reported primal-dual gap is 9.93e-13; thus the upper gain ceiling differs from
the achieved gain by at most twice that gap, subject to numerical correctness.
All ten matched-world ceilings remain below .001547, far below .005.
[Machine-readable ceilings](simplex-guard-ceiling.json) include an extra 1e-10
reporting cushion. These are floating-point numerical certificates, not formal
interval-arithmetic proofs; the large gap to the requirement makes ordinary
roundoff an implausible explanation of the failure.

This exhausts changing only the projection target under THIS guard for the
matched-world gain requirement. It does not prove the original population
nonharm requirement impossible. The universal per-history/per-outcome guard is
strictly stronger than that requirement. Changing observations, obtaining
reliable information about the likelihood, or allocating risk across histories
remains unexplored here.

## Verification

Full byte-exact replay passes. All three original controls and probability
masses match the preceding mismatch artifact in every world. Alternate
conditional-risk scoring agrees on all 3200 Brier values, max error6.11e-16,
and all gate decisions agree. This is scoring-algebra verification, not an
independent inference implementation.

All 204800 vertex-regret checks pass within floating-point tolerance; maximum
computed regret .010000000000000231. At most71 projection cycles were needed,
no convergence failures. 37000 of51200 report-vector evaluations alter the
optimistic proposal. These unweighted enumeration counts are not activation
rates. No serving or reference latency benchmark was performed.

Artifacts: [full results](simplex-guard-noise.json),
[verification](simplex-guard-noise-verification.json),
[ceiling derivation script](simplex-guard-ceiling.mjs).

## Next directions

Do not relax the original .01 population harm or .005 gain requirements.
Investigate a justified likelihood uncertainty envelope, or a population-risk
constraint that can spend less risk on common harmful histories and more on
useful ones. Any finite envelope needs explicit coverage limits; estimated
coverage requires untouched evidence and uncertainty accounting. The universal
guard can be an honest fallback, but a fallback alone is not the research win.

No production or whitepaper modifications or publication. All seven research
directions remain open.

