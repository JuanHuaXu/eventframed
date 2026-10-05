# Finite spike-and-slab component: algebra verified, no quality claim

## What changed

A research-only `_test.go` component implements a finite Gaussian slab and
point-mass exclusion prior, with a conditional Gaussian intercept. This is
distinct from the failed Gamma-precision model and from the regularized
horseshoe. The prior inclusion probability and slab variance remain explicit
component inputs; no values have been selected on the research tape.

Derivation and primary-source attribution are in
`research/spike-slab-component-derivation.md`. The proper unit-variance
intercept changes the collapsed formula relative to the source's intercept
convention; its covariance with the selected coefficients is retained.

## Checks performed

- Collapsed lower bound matches an uncollapsed expected-likelihood and
  conditional-intercept-KL calculation for nonoptimal and updated states.
- Fixed-xi coordinate updates increase the bound. Perturbing each updated
  inclusion log-odds, mean or log variance does not improve that coordinate.
- Updating xi from the joint moments and reprofiling the intercept increases
  the bound across the tested states.
- Tests cover 1/7/64 labels; 1/3/9 feature subsets; three prior inclusion
  probabilities and three slab variances. These are algebra fixtures, not
  a quality hyperparameter search. A separate full 255-feature sweep passes.
- A single-feature surrogate update agrees with independent 131,072-panel
  Gaussian quadrature for inclusion, slab mean and slab variance. This is
  the fixed-xi quadratic surrogate, not the exact logistic posterior.
- Explicit enumeration of all eight inclusion states for a three-feature
  mixture agrees with joint predictive moments for eight query states.
- Label complement symmetry, input immutability, and malformed prior,
  feature, xi and factor rejection checks pass.

`go test ./internal/observationlearners -run '^TestSpikeSlab' -race -count=1 -v`
passes in 1.586 seconds package time. No failed experimental run or behavioral
patch was needed in this component. These finite tests do not prove arbitrary
posterior calibration, global optimization, or future-data safety of a fitter
that has not yet been implemented.

## Performance

`BenchmarkSpikeSlabReferenceSweep`, Apple M4, darwin/arm64, three repeats of
10 operations: 12,027,367 / 11,229,662 / 11,298,812 ns/op; about 2,245,315-
2,245,869 allocated bytes/op, 1,108 allocations/op. Package time 0.751 seconds.

One operation constructs the dense fixed-xi profile for 64 labels and 255
features, performs a coordinate sweep, and evaluates the resulting bound.
It is NOT a converged fit, xi cycle, prediction or serving request. Dense
Gram construction costs O(n*p^2), storage O(p^2+n*p); coordinate copies also
allocate throughout the sweep. This deliberately transparent reference is
not the intended runtime path. Maintaining X*E[beta] should allow an O(n*p)
sweep/profile without changing equations; equivalence must be tested.

## Remaining work

Implement and verify the matrix-free sweep, then a bounded fixed-prior fitter.
Freeze prior and initialization choices before trajectory quality collection.
Do not reuse the Gamma update or select a slab scale from consumed outcomes.
The predictive law must integrate the spike-and-slab mixture. A Gaussian with
the same first two moments is not that law and must not be silently substituted.
As-of adapter, mixture integration, quality pilot, full quality experiment,
and runtime integration remain open. All seven research goals remain OPEN.

Numerically saturated inclusion probabilities currently fail closed rather
than silently clipping a coordinate optimum. Any later log-domain boundary
handling needs explicit tests. No production, paper, install, commit or push
changes were made; no experimental processes remain running.

SHA-256:
- Component: `3710cd0893c837fe6dfdbec9de192e0f25499ce4ef556458b9ceb3be0a3d9a60`.
- Derivation: `b0726985061875e528dd8719d3e124beab8345ba50c2c8cde5401d98f34042c3`.
