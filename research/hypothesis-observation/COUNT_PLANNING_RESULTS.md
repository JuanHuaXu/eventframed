# Exact six-query planning: two-query lookahead reaches the model optimum

The [frozen diagnostic](COUNT_PLANNING_PROTOCOL.md) solves all48048 count states
(3003 per initial report pattern) under the SAME interval-aware joint prior.
It does not change the model or the16-credit acquisition budget.

Renewal order is removable in this finite iid-or-root-copy model: initial roots
and per-source zero/one counts determine the likelihood. Ordered/permuted
history checks agree; no multinomial-count factor is inserted. This sufficiency
does not apply automatically to arbitrary temporally dependent event sources.

## Exact model-prior expectations

Lower values are better. Full planning minimizes the SUM of six post-query
Bayes risks, not final loss and pre-query area separately.

| Policy | Post-query risk sum | Learning-area risk | Final risk |
| --- | ---: | ---: | ---: |
| Fixed | 2.571583407 | .443522754 | .391168189 |
| Random | 2.618916806 | .448555757 | .408303570 |
| Report entropy | 2.359463943 | .415166784 | .349184547 |
| One-query | 2.346534416 | .413712276 | .344982068 |
| Two-query | 2.339098834 | .412930375 | .342237889 |
| Full six-query | 2.339098834 | .412930375 | .342237889 |

Two-query and full planning have bit-identical initial expected scores for
every one of the16 initial root patterns. Their actions differ at12 of20592
nonterminal states, but the largest two-query action Bellman excess is
4.44e-16; a six-stage accumulation bound from these numerical residuals is
2.67e-15. Do not claim identical policies or symbolic exact equality: this
is finite floating-point verification with roundoff-scale differences.

Full planning improves modeled final risk by .002744 over one-query and
.006947 over entropy. Area gain is .000782 versus one-query and .002236 versus
entropy. These are expectations under the declared prior, not the fresh
rollout distribution or independent real-world observations.

## Verification

- Count sufficiency:90 likelihood/forecast comparisons, max error5.55e-17.
- Existing depth1/depth2 cost comparisons:36; invalid count case rejected.
- Initial root masses sum to1; every root enumerates exactly3003 states.
- Backward solver checks82368 action inequalities.
- Independent forward occupancy on saved beliefs checks288 policy values and
  672 layer normalizations, max value disagreement1.29e-14.
- Separate verification recomputes20592 nonterminal Bellman minima.
- Full output replay is byte-exact.

[Complete states and policies](count-planning-exact.json),
[verification](count-planning-verification.json),
[two-query excess check](count-planning-two-gap.json),
[planner](count-planning.mjs).

The forward verification is independent of the backward recursion but uses
the saved belief model. It is not an independent likelihood implementation or
formal interval-arithmetic proof. Earlier direct-quadrature tests provide
separate finite checks on inference.

## Consequence

Within this model, budget and summed-risk objective, increasing planning depth
beyond2 does not offer a meaningful improvement. A larger looping budget alone
cannot explain away the failed fresh per-regime screens. Those failures remain:
20/108 in the recent two-query rollout,19/84 in the preceding one-query rollout,
and the impossible900-gate fixed-observation screen is still failed.

The model-prior average is not the same requirement as protected improvement
in each real or simulated regime. Source prior adequacy, source dependence,
uncertain noise and the mismatch between averaged and subgroup objectives
remain live leads. Per-regime oracle bounds or a robust regime-sensitive
objective would be more discriminating than another lookahead-depth sweep.

State compression makes a finite reference tractable; no timing benchmark or
production-scale complexity result is claimed. Unknown hypotheses, actual
agent tasks, open-ended temporal processes and serving integration remain
untested by this diagnostic. No production changes, whitepaper promotion or
publication. All seven research directions remain open.

