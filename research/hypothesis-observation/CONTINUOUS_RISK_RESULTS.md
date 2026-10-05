# Continuous noise diagnostic: a bounded partial result

[Protocol](CONTINUOUS_RISK_PROTOCOL.md), [coefficients and envelopes](continuous-risk.json),
[verification](continuous-risk-verification.json).

Constructed degree10 Bernstein population-risk polynomials from all48048 count
states for four frozen policies. Direct likelihoods and squared forecast loss,
with policy path multiplicities, give these polynomials without simulating
trajectories or adding observation samples. Ten subdivision levels tighten the
coefficient envelopes on each of three noise intervals.

These are ordinary floating-point computations. The coefficient convex-hull
bound is mathematically valid in exact arithmetic, but the implementation does
not provide outward-rounded error enclosures. Therefore this is stronger
numerical evidence than checking a grid, NOT a rigorous real-arithmetic
certificate. Near-zero bounds must not be promoted to proven strict signs.

## Findings

Across noise [.10,.30], versus normalized entropy, eight masks have positive
area-gain lower bounds and seven have identically zero stored gain coefficients
(excluding all-copy mask15). Positive masks are0,2,4,6,8,10,12,14. The smallest
positive lower bound is .0000174358273; the largest upper bound is .000343163177.
The unchanged odd masks all copy source0, a direct target-bit source. This is
a structural lead, not proof that those regimes lack acquisition headroom.

| Noise interval | Control | Lowest final-gain bound | Masks with positive area lower bound /15 |
| --- | --- | ---: | ---: |
| [.10,.30] | Random | -.00181944110 | 12 |
| [.10,.30] | Archived entropy | -.00543749824 | 8 |
| [.10,.30] | Normalized entropy | 0 | 8 |
| [.05,.10] | Random | .00005041166 | 15 |
| [.05,.10] | Archived entropy | -.00050657723 | 11 |
| [.05,.10] | Normalized entropy | 0 | 8 |
| [.30,.35] | Random | -.00239626132 | 12 |
| [.30,.35] | Archived entropy | -.00666909518 | 8 |
| [.30,.35] | Normalized entropy | 0 | 4 |

All numerical final bounds remain above the -.01 allowance. No broad success
follows: area dominance still fails against the original controls, and seven
normalized-entropy area comparisons remain tied. On [.30,.35], masks8,10,12,14
have sign-changing or negative area behavior, consistent with the prior direct
counterexamples at.35. Worst area lower bound versus normalized entropy is
-.000210664414. A negative coefficient bound alone is not a counterexample;
the separately evaluated .35 outcomes supply the actual negative witnesses.

## Verification

- Byte-exact replay and frozen source hashes.
- 330 known-polynomial, subdivision and degree-elevation checks; maximum error
  2.23e-16.
- 1408 archived on-grid/off-grid score comparisons using the separate direct-
  likelihood evaluators; maximum discrepancy1.49e-14.

These comparisons validate the construction numerically, not rounding bounds,
source authenticity, model adequacy or actual-agent performance.

## Next work

For an exact protection claim, add directed-rounding or rational verification.
For a useful learning rescue, investigate why copying source0 prevents earlier
safe action changes, and whether a less restrictive global risk allocation can
use the alternate source relation without harming controls. Do not treat a
numerical certificate upgrade as solving the remaining learning failures.
All seven directions remain open; no production or paper changes.
