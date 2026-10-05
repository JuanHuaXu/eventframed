# V70 Prior-Responsiveness Algebra

**PASS finite algebra and independent readback; NOT a stream rescue.**
All seven whole goals remain open. No new empirical outcomes, future seeds,
agent tasks, production state or whitepaper were accessed or changed.

[Prospective derivation](../../research/rate-prior-v70-direction.md),
[checked results](../../research/rate-prior-v70-recheck/results.json),
[independent audit](../../research/rate-prior-v70-audit/results.json).

## Tradeoff

The current member prior has .8 mass at its baseline and .2 on the finite
strength1 urn law. At unchanged first moments, removing that spike increases
the first noisy observation's posterior-mean movement by5x; its stationary
one-observation excess Brier risk increases25x. Removing the spike while
using strength2 instead gives3.49206349x movement and12.19450743x excess risk.
The ratios hold across the seven declared baselines and three fixed noise rates.

These are relative excess-risk multipliers over a perfectly correct baseline,
not total Brier multipliers or forecasts of daemon accuracy. The calculation
concerns one local update under fixed noise. It does not uniquely diagnose
V68/V69's mixture failures, model the full stream, or establish that these
priors would fail every future controlled experiment. A broader prior alone
has not solved the recovery-versus-stationary-protection requirement.

## Verification

2,079 preflight checks compare prior ratios against independent Polya urn
dynamic programming, means/second moments, noisy posterior branches and tower
identity, three reset hazards, and explicit shared-Y paired measurements.
Maximum checked defect7.17648e-13 under the unchanged2e-12 tolerance.
42 impossible zero-noise pair branches are excluded.42 negative controls
detect the erroneous independent-Y double counting of a repeated measurement.

Independent readback checks399 quantities using direct urn second moments,
with maximum defect2.55351e-15, and rejects all nine altered-result controls.
Source copies, freezes, output hashes and terminal completion records are
retained. This is not a Go integration, a benchmark, an empirical confidence
interval or an Anti-Pigeon certificate.

## Preserved Checker Failure

The original script stopped at its ratio comparison:24.999999999999968 versus
24.999999999997744. Dividing subtracted posterior movements amplified their
rounding error past the absolute tolerance. The new script compares the
directly computed variance ratio instead; all posterior-branch checks and the
tolerance remain unchanged. The original source, freeze and
[failure record](../../research/rate-prior-v70-preflight/failure.json) remain
unchanged. That record explicitly identifies its stderr capture as transcribed
from the tool output rather than an independently saved raw log.

## Next Boundary

Avoid an unconditional fast-prior substitution. The earlier
[V34 dispersion study](mmm-dispersion-v34-results.md) showed useful low-variance
protection from coherent concentration learning but retained nonlinear and
shape failures. The [V35 shape study](mmm-shape-v35-results.md) repaired some
specific shape cases, not the full goal. A future dynamic/noisy integration
must retain independent member rates, a declared joint hypermodel, original-
position pair replacement, and all quality, recovery, memory and total-cost
gates. No hierarchy is an external guarantee or automatic authority to pool.
