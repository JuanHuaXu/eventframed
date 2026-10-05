# Paired-risk v73: evaluation refinement, not a model rescue

Implemented and checked a conservative paired excess-miss bound that counts
both harmful and beneficial discordances. The derivation and finite validation
support using it in a future predeclared experiment. The application to v72 is
**post-hoc diagnostic only**; v72 remains failed under its frozen protocol, and
its primary speed shortfall is unchanged. No new outcome data were collected.

## Finding

For v72's late-change confirmation cell, corrected acquisition misses47 cases
the control detects and rescues79 cases the control misses. Net excess misses
are -32/512=-6.25 percentage points. The old harmful-only upper bound is13.053%.
The new paired upper bound is0.596%, below the numerical2% threshold.

The design late-change cell has37 harmful and70 beneficial discordances: its
bound changes from10.778% to-0.190%. All twelve diagnostic paired bounds are
below2%, but this does not retroactively pass a predeclared rule. The new bound
is not uniformly tighter: for zero discordances its upper is1.277%, compared
with1.065% for the old harmful-only bound, because it spends error probability
on three constituent bounds. It was not formed by taking an unadjusted minimum
of several methods after looking at the sample.

This distinction matters: the v72 late-case failure was an inability of its
conservative certificate to establish non-harm, not evidence of worse aggregate
deadline reliability. Its ~5% primary detection gains still miss the10% target,
and its weak-change sensitivity and actual MMM integration remain unresolved.

## Derivation and checks

Let s be the probability of either discordance and r the harmful fraction among
discordances. Excess risk is s(2r-1). A binomial confidence interval for s and
conditional binomial upper bound for r give a rectangle with joint coverage
at least1-alpha via a three-way union bound. Maximize s(2r-1) over this rectangle.
The detailed sign-dependent endpoint rule is in the protocol. Independence
between the three intervals is not assumed. This is a fixed-sample result,
not an anytime guarantee under repeated monitoring or adaptive study selection.

Building block: [Clopper & Pearson (1934)](https://doi.org/10.1093/biomet/26.4.404).
The specific rectangular paired construction is our derivation, not claimed to
be their original paired-outcome method or an optimal confidence bound.

Tests enumerate every possible observed (harmful,beneficial) count pair at
sample sizes1,4,8,16,32 over the frozen probability grid and alpha=.05,.01.
All660 parameter cells have unit probability mass within numerical tolerance
and noncoverage below the nominal error budget. The maximum observed ratio of
noncoverage to that budget is0.32913. This checks floating-point implementation
on the declared finite grid; the proof, not that grid, covers other parameters.
Analytic binomial special cases, interval ordering, zero-discordance uncertainty,
input rejection and deterministic artifact reproduction also pass.

Boundary audit found one implementation error before finalizing: validating
alpha only after division by3 accepted invalid original confidence budgets
such as1.5. A failing regression was added, the original alpha is now checked
before allocation, and all tests pass. Numerical outputs for the valid study
inputs are bit-identical before and after that fix.

## Artifacts

- Final analysis: `mmm-paired-risk-v73-checked.json`, SHA256
  `d7918bfff3de3ac2c207e246a0880dd618306c3fcb8e46475e9d81c4631f5e29`.
- Earlier `mmm-paired-risk-v73.json` is retained as a superseded pre-audit
  output. Its source hashes precede the input-validation repair; do not use it
  as the current-source verification artifact. Its numerical cells match final.
- Consumed v72 source SHA256 remains
  `1fce6db98b38b732519073509e12f10fa99d7dd59f9719ddfa1d4d977b5e61e0`.
- Analysis, test and protocol hashes are in the checked output, with all13
  v72 source hashes verified before analysis. Replay equals the checked output.
- Tests: `node --test research/paired-risk-v73.test.mjs`, two tests passed,
  including full sample-space enumeration over660 parameter cells (~70ms total).
- Analysis: `node research/paired-risk-v73.mjs <unused-output-path>`.

This is offline evaluation code only. No serving path, research detector,
original v72 artifact/protocol, whitepaper, production system or remote changed.
Next use the paired method only in a freshly frozen protocol, while pursuing
the remaining acquisition/stopping-time and integrated-learning gaps. All seven
roadmap directions remain incomplete.
