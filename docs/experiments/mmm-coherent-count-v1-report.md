# Coherent count component: valid model, preliminary quality headroom

An isolated test-only alternative now implements a single uniform-input,
symmetric-outcome Dirichlet prior of total mass2. It accumulates full-cell
counts, then marginalizes to all19683 partial states. Original count models
and all earlier experiments remain unchanged.

## Mathematical Checks

Literal full-state enumeration agrees exactly with every partial forecast in
the counterexample fixture. All59049 parent/one-bit partitions have additive
mass and positive-outcome mass. The old contradictory two-positive-example
case now gives .75 at the root and for both bit0 values. Neutrality, invalid
inputs, immutable fitted ownership and bounded probabilities pass.
These checks establish the component's tested algebra, not calibration or
generalization in an unknown environment.

## Fixed-Mask Quality

192 independent fits: two input laws, bit/parity4 targets, n64/128/4096,
16 fits per cell. Expected Brier is computed exactly over the finite population
and .05 label noise, on four predeclared masks0/1/31/511. Both estimators use
identical past samples. Averages below use mask31, which includes all relevant
fields in these specified generators:

| Input | Target | Labels | Old | Coherent |
|---|---|---:|---:|---:|
| Uniform | Bit |64|.124149|.089545|
| Uniform | Parity4 |64|.126989|.097105|
| Uniform | Bit |128|.086274|.066342|
| Uniform | Parity4 |128|.085791|.063827|
| Dependent | Bit |64|.088261|.068208|
| Dependent | Parity4 |64|.082011|.059277|
| Dependent | Bit |128|.063253|.055755|
| Dependent | Parity4 |128|.061565|.053653|

Root predictions match exactly. Full-input forecasts also improve in these
low-noise fixtures, although sparse full cells remain poor predictors. At n4096,
mask31 differences are tiny. Uninformative mask1 cases generally worsen a
little, up to .000395 mean Brier here. Every mask and result is retained in
the summary, including regressions; no automatic model promotion follows.

Important omissions: no null-label quality fixture, noise sweep, adaptive
retrieval, changing regime, delayed evidence, joint expert mixture or actual
agent task. The known useful mask is evaluator design, not information given
to an observer. Next test null/high-noise robustness before integrating the
candidate into observation selection. Less smoothing may overfit those cases.

## Performance and Verification

- Race-enabled component checks PASS1.486s; vet PASS.
- Quality collector completed all192 fits in.12s (test),.449s package.
- Independent summary checks dimensions, unique seeds, finite risks and root
  parity. Summary including six captured source files replays byte-identically.
  Source capture is post-collection, not a separately timestamped preregistration.
- Fit64 microbenchmark, Apple M4/darwin arm64, three300ms repetitions:
  72093/72135/73775ns per fit,319488-319489 bytes and one allocation.
  This is component fitting, not daemon serving or a matched speedup claim.

The representation remains exponential in dimension: O(3^d) space and a
descending marginalization pass bounded by O(d*3^d); this nine-bit experiment
does not justify an unrestricted high-dimensional implementation.

Raw:`mmm-coherent-count-v1.json`.
SHA256:`8de76fe379cfe1bb9219880ae94fb4b94f48dccf800ba343628f2d2c7db8a7c2`.
Contract:`mmm-coherent-count-v1-contract.md`.
Summary/replay:`mmm-coherent-count-v1-summary.json` and corresponding replay.
No production, whitepaper or remote changes. All seven goals remain open.
