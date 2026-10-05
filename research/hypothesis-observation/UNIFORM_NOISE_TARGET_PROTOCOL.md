# Frozen uniform-noise posterior target

The interval line guard passed873/900 gates but missed27 gain requirements.
Test one change: retain that identical guard and fixed-noise local baseline,
but replace its fixed-.20 optimistic target with the all-genuine posterior
predictive under a uniform noise prior on [.10,.30]. Actual noise and mask
never enter this calculation. Do not tune the prior after reading results.

For each history, use the mask0 joint likelihood coefficient arrays from the
noise envelope. Each Bernstein basis term integrates to1/(degree+1) on [0,1],
so integrating each joint class/history mass amounts to taking the arithmetic
mean of its coefficients. Normalize these integrated joint masses ONCE.
Averaging normalized conditional forecasts at noise values is not equivalent.

The target conditions on the optimistic all-genuine mechanism. This does not
authenticate reports or change the source posterior. The guarded final output
is a constrained decision forecast, not claimed to be an ordinary posterior.
The fixed baseline, epsilon=.01, envelope, allocations and all900 gates remain
unchanged. No new data or fresh-confirmation claim: this is a rescue on the
consumed noise stress family.

Verify the target by independent six-node Gauss-Legendre integration of direct
joint likelihoods (degree at most10); validate quadrature on monomials0..11.
Replay the complete run, compare all original controls and recompute Brier
through the conditional-risk identity. Report every failed gate, including
cases where both forecasts improve but fail the predeclared minimum.
No solver tolerance changes, production work or whitepaper promotion.

