# Off-grid budget policy: interpolation holds, extrapolation has area harms

[Protocol](RISK_BUDGET_OFFGRID_PROTOCOL.md), [results](risk-budget-offgrid.json),
[verification](risk-budget-offgrid-verification.json).

The candidate and guard were frozen at their prior hashes. Six new evaluation
noise levels were not supplied to the policy compiler: .125/.175/.225/.275
inside its range, and .05/.35 outside it. All16 masks were evaluated at each.
This tests noise transfer within the same finite source-mechanism family, not
new semantic tasks or real-world data. The candidate was not tuned afterward.

## Scores

All288 final nonharm checks pass. Positive-area checks pass183/270;87 still
fail, so the full screen remains FAIL. Mask15 is excluded from area conditions
under the unchanged protocol.

| Range | Control | Final nonharm | Positive area | Area harms | Area ties |
| --- | --- | ---: | ---: | ---: | ---: |
| Interpolation | Random | 64/64 | 56/60 | 4 | 0 |
| Interpolation | Archived entropy | 64/64 | 37/60 | 23 | 0 |
| Interpolation | Normalized entropy | 64/64 | 32/60 | 0 | 28 |
| Extrapolation | Random | 32/32 | 27/30 | 3 | 0 |
| Extrapolation | Archived entropy | 32/32 | 19/30 | 11 | 0 |
| Extrapolation | Normalized entropy | 32/32 | 12/30 | 4 | 14 |

All counted positives exceed1e-12; counted ties are exactly zero in the output.
Against normalized entropy, the maximum interpolated final gain is .0021722897
and maximum area gain .0003388467. No interpolated area harm was found at these
four points. This does not establish the inequality throughout the interval.

All four extrapolated area harms against normalized entropy occur at noise.35:

| Copy mask | Final gain | Area gain |
| --- | ---: | ---: |
| 8 | .0003778731 | -.0000234508 |
| 10 | .0003019802 | -.0002106644 |
| 12 | .0003778731 | -.0000234508 |
| 14 | .0003019802 | -.0002106644 |

These are small but genuine losses under the strict learning-speed condition.
Positive terminal gain does not cancel intermediate prediction harm. They do
not contradict the original guard's finite-family induction: noise.35 was never
covered. They do refute an extension of area protection to this extrapolation.

## Verification

Full replay is byte-identical. Six frozen source hashes and all three compiler
statistics match; zero budget reproduces20592 actions. There are8064 world/
policy/stage mass checks,33792 direct off-grid likelihood comparisons (maximum
error1.39e-17), and360 separate backward-score comparisons (maximum5.60e-14).
Invalid-budget and tie-rule tests also pass. These are numerical checks, not
interval-arithmetic certificates or independent validation of the source model.

## Next lead

Stop treating more noise-grid points as a continuous guarantee. For this fixed
finite policy, derive polynomial population-risk differences in the common
noise parameter and check an interval envelope. That can establish where the
partial protection actually holds or provide another counterexample. It will
not repair the35 on-grid area ties, nor prove the policy discovers useful
real-world observations. More expressive trajectory-budget allocation remains
open. No production or whitepaper change; no research direction is complete.
