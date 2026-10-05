# Canonical tie-policy comparison: not a quality rescue

[Protocol](TIE_POLICY_PROTOCOL.md), [evaluation](tie-policy-evaluation.json),
[verification](tie-policy-verification.json).

This is a diagnostic on 80 consumed finite source regimes, not new confirmation.
All forecasts and model evidence masses remain frozen. The only change is choosing
the first observation type within 1e-12 of the minimum objective, consistently
for one-step, two-step, full-horizon and entropy policies.

## Findings

The rule changes actions in 92/84/86/832 of 20592 nonterminal states for
one/two/full/entropy respectively. The maximum full-horizon model-prior cost
excess is 4.44e-16. There are 82368 normalized branch checks.

Across all 80 worlds, final Brier and learning-area scores for one/two/full are
bit-identical to their archived counterparts. This does not assert identical
paths: changed actions may be unreachable or have equal aggregate effects.
Entropy changes: maximum absolute final Brier difference is 0.0054374982 and
maximum absolute learning-area difference is 0.0068220973.

Both tie_two and tie_full have the following failures:

| Control | Final nonharm failures / 80 | Positive-area failures / 75 | Worst final gain |
| --- | ---: | ---: | ---: |
| Random | 0 | 9 | -0.0000095531 |
| Archived entropy | 4 | 32 | -0.0157209502 |
| Tie-normalized entropy | 1 | 32 | -0.0102834519 |

The final allowance remains 0.01; positive area gain remains required outside
the all-copy mask15. The remaining final failure against normalized entropy is
noise0.30, mask3: both direct target sources copy their initial reports. Its
area gain is -0.0120699635. Four final failures against archived entropy remain
at masks3/7 and noise0.25/0.30. None are erased or renamed successes.

The experiment separates two issues: tie conventions affect the control, but
the planning weakness survives a shared convention. Further planning depth or
numerical tie cleanup alone is not a demonstrated rescue.

## Verification

- Complete byte-exact replay.
- All 480 archived-policy/world score objects exactly preserved.
- Four tie-rule tests, including a non-tie and invalid-input rejection.
- 16896 direct likelihood comparisons; maximum discrepancy 1.39e-17.
- 300 separate backward-score comparisons in six regimes; maximum discrepancy
  2.98e-14. The main evaluator instead propagates forward path multiplicities.
- 5600 total world/policy/stage probability normalization checks.

These checks validate finite computation and reproducibility, not the adequacy
of the source family for actual agents. No runtime or production changes.

## Next discriminating lead

With ties fixed, compare a source-regime-aware observation objective against the
prior-average objective while holding the predictive law fixed. A candidate
could minimize worst-regime continuation risk or control regret to the baseline
acquisition policy, rather than only prior-average target uncertainty. It must
not observe the actual regime and must retain every failed regime as a control.
Any design using this grid needs fresh independent validation before promotion.
All seven research directions remain open.
