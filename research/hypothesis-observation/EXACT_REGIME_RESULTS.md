# Exact regime evaluation: hidden subgroup tradeoffs confirmed

The [frozen diagnostic](EXACT_REGIME_PROTOCOL.md) evaluates saved compiled
count-state policies in80 regimes: five noise values and all sixteen copy masks.
Actual noise/mask enters scoring only; all forecasts and action rules are frozen.

## Findings

Two-query and full policies each fail final-loss protection against entropy in
four regimes. The eight failed comparisons count the same four regimes twice,
once per policy; they are not eight independent discoveries.

| Noise | Copy mask | Copied source types | Final Brier harm vs entropy | Area harm vs entropy |
| --- | ---: | --- | ---: | ---: |
| .25 | 3 | 0,1 | .012697 | .015720 |
| .25 | 7 | 0,1,2 | .012674 | .015720 |
| .30 | 3 | 0,1 | .015721 | .018892 |
| .30 | 7 | 0,1,2 | .015425 | .018892 |

Types0 and1 directly report the target bits; type2 reports a context bit.
The harmful patterns copy both direct target sources. These are exact finite
population expectations, not confidence-interval misses. They show that
model-prior optimal observation can violate per-regime protection.

For EACH candidate (two/full):
- Versus random:0/80 final nonharm failures;9/75 positive-area failures.
- Versus entropy:4/80 final nonharm failures;32/75 positive-area failures.
- Fully copied mask15 is excluded from the positive-area requirement, as
  predeclared, but remains in final nonharm checks.

Across both candidates,82 area comparisons fail; all are negative by more than
1e-12, not numerical ties. Most regimes still improve, but uniform superiority
is not established. No new confidence-screen pass is inferred.

## Relationship to earlier tests

The recent rollout grid used masks0,5,10,15, so it omitted the newly confirmed
harmful masks3 and7. On that old grid, these COMPILED policies have no exact
final nonharm failures but still have six positive-area failures across both
candidates and controls. This helps motivate broader evaluation.

Do not equate these compiled policy values with the precise expectations of
earlier ordered-history implementations: floating-point ties can select
different actions. The current policies are explicitly frozen in the count
artifact. Neither the earlier sampling failures nor their source code is
retroactively changed.

Full planning remains optimal for its declared prior-weighted objective.
That does not imply optimality or protection in every actual source regime.
The result does not show that more planning depth would fix these failures.

## Verification

The scorer propagates action-only path multiplicities, then weights each
count-state by its actual joint history/class likelihood. This counts multiple
report orders without inserting a second multinomial factor. All3360
policy/stage probability normalizations pass.

Full output replay is byte-exact. Independent ordered-product checks agree on
16896 joint likelihood values within1.39e-17. A separate backward conditional-
probability evaluation in six representative regimes agrees on180 score
quantities within2.98e-14. These include harmful, genuine, mixed and fully copied
regimes. The action table stays frozen throughout.

[Full80-regime table](exact-regime-evaluation.json),
[verification](exact-regime-verification.json),
[scoring model](exact-regime.mjs).

## Next research

Investigate source-prior adequacy and regime-sensitive objectives, keeping all
sixteen copy masks visible. A controlled prior/policy ablation or robustness
constraint should separate better forecasts from better query selection.
Check numerical tie and order invariance before attributing differences between
compiled and ordered-history controllers. Do not tune only masks3/7 and call
that independent validation.

This remains a specified finite source/noise family, not arbitrary poisoning,
real-source authentication, actual-agent efficacy or continuous-noise coverage.
No production change, performance benchmark, whitepaper promotion or
publication. All seven research directions remain open.

