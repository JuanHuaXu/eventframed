# Uniform-noise posterior target: no new gates rescued

The [frozen candidate](UNIFORM_NOISE_TARGET_PROTOCOL.md) marginalizes common
channel noise under a uniform [.10,.30] prior when constructing its optimistic
all-genuine target. The local baseline and interval line guard are unchanged.

Overall873/900 gates pass (FAIL), with exactly the same27 failures as the
fixed-noise interval line guard. All800 protection and50 false-confidence
requirements pass. Seven genuine gains fail at .20, and all ten fail at each
of .25 and .30. No gate changes from fail to pass or pass to fail.

| Actual noise | Pass /180 | Worst population harm | Genuine gain range |
| --- | ---: | ---: | ---: |
| .10 | 180 | -.000972 | .006515 to .007021 |
| .15 | 180 | .001679 | .005603 to .006392 |
| .20 | 173 | .004367 | .004077 to .005495 |
| .25 | 170 | .007043 | .001819 to .004184 |
| .30 | 170 | .009701 | -.001042 to .003254 |

Across800 population cells, candidate Brier changes versus the prior line guard
range from -.002603 (improvement) to +.000531 (harm). Some scores improve without
meeting the original .005 gain threshold. This does not support broad dominance
or a claim that the remaining gap is merely an implementation error.

## What is and is not inferred

The target integrates joint class/history likelihoods before normalizing.
It is a posterior predictive under a declared uniform noise prior and the
optimistic all-genuine mechanism, NOT evidence that the reports are genuine.
The final guarded decision is not claimed to be an ordinary Bayesian posterior.
The prior has not been estimated or validated on real data, and the guard's
continuous-noise protection still assumes the declared measurement family.

The ten reports can update the noise posterior, but the results do not establish
that this inference identifies actual noise reliably. Nor do these results prove
that every noise-aware target or a joint noise/source learner must fail.

## Verification and artifacts

Independent six-node Gauss-Legendre integration of direct report likelihoods
matches732 target probability values within5.55e-16. The quadrature itself passes
monomials of degrees0 through11, max error1.39e-16. Likelihood products have
degree at most10. Two target negative-input checks and the existing107100
envelope checks pass.

Full replay is byte-exact. Original controls and masses match in every world.
All3200 Brier values agree with the alternate conditional-risk identity,
max error6.11e-16; all gate decisions agree. The maximum computed coefficient-law
regret is .010000000000000231. Tests are finite numerical checks, not a formal
rounding-error proof or fresh empirical confirmation.

[Full results](uniform-target-noise.json),
[verification](uniform-target-noise-verification.json),
[target implementation](uniform-noise-target.mjs).

## Next meaningful change

Per-history conditional guards are stronger than the original population
protection requirement. Test allocating the unchanged .01 population risk
budget across report histories, under a predeclared family of generating laws.
Keep all original gain and false-confidence gates. A finite reference may
precompute a history-to-correction rule without giving it the actual world;
do not present that exhaustive lookup as a scalable production implementation.

Any learned risk allocation needs separate, honest validation under model
mismatch and new observation structures. The prior closest-point numerical
failure also remains unresolved and preserved. No production changes,
performance claims, whitepaper promotion or publication. All seven goals open.

