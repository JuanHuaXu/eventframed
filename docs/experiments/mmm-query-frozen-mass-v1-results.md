# Fixed-support mass substitution and its limitation

## Diagnostic Result

The [frozen protocol](mmm-query-frozen-mass-protocol.md) completes all 2688
consumed records, with no new posterior fits. For the 1344 delayed histories,
teacher-weighted population Brier is:

| Selector | Frozen64 risk | Paid queries |
| --- | ---: | ---: |
| No query | 0.170199 | 0 |
| Random | 0.168501 | 1344 |
| Entropy | 0.167964 | 1344 |
| Joint8 | 0.168333 | 1344 |
| Teacher-weighted branch-risk oracle, paid | 0.165850 | 1344 |
| Model-weighted branch-risk oracle, paid | 0.168979 | 1344 |
| Neutral-weighted branch-risk oracle, paid | 0.167325 | 1344 |
| Teacher-weighted oracle with abstention | 0.165679 | 1085 |
| Model-weighted oracle with abstention | 0.169892 | 174 |
| Neutral-weighted oracle with abstention | 0.168604 | 392 |

Model and teacher weights choose different paid queries in 1115/1344 histories.
The probability MSE, equally weighted across pools, is 0.045554. Phase1 retains
the same qualitative result: entropy 0.167888, teacher-paid oracle 0.165627,
model-paid oracle 0.168827, neutral-paid oracle 0.167156. These oracle losses
are unavailable online; neutral's lower aggregate is not a deployable rescue.

## Important Interpretation Correction

After seeing that the model-weighted oracle abstained on **every** frozen63
history, we investigated its algebra. This check was motivated by the result,
not a predeclared efficacy screen.

Let f0 and f1 be the forecasts after the two query answers, p the model's answer
probability, and f=(1-p)f0+p*f1 the unchanged forecast. Let q(x) be the fixed
teacher's future outcome probability, identical in the two counterfactual
evaluations. For squared probability loss on any fixed target distribution:

```math
(1-p)R_q(f_0)+pR_q(f_1)-R_q(f)
=p(1-p)\,\mathbb E_x[(f_1(x)-f_0(x))^2]\ge 0.
```

The Bernoulli irreducible term q(1-q) cancels. Forecast aging multiplies each
input's squared difference by 0.99^(2*age). Thus the model-weighted fixed-world
oracle cannot prefer an update over the coherent frozen63 baseline: it sees
only added forecast variance. The observed zero-query choice is structural,
not a standalone demonstration of bad calibration. Frozen64 abstention is not
subject to the same baseline identity, because its restored label changes f.

All 9277 paid candidates satisfy the sampled-target variance identity, maximum
error 4.34e-15. The integrated population surplus is also nonnegative throughout.
The mixture coherence on all512 inputs was independently checked in the prior
coverage experiment.

The mass substitution experiment remains a valid sensitivity calculation for
true risk under a fixed teacher, and its regret/distortion bounds remain true.
However, replacing the query marginal while holding the teacher's future law
fixed is **not** a coherent Bayesian joint-model intervention. A Bayesian learner
expects the query answer to change its belief about the future outcome law.
These runs therefore do not establish that answer-probability error alone
explains the heuristic failure, nor that calibrating that marginal is sufficient.
This qualification also applies to interpreting the earlier publication-C
mass-substitution diagnostic; its numerical results are not retracted.

## Verification

- Existing selector contracts pass: 625 identity/relabeling cases and 3125
  regret/order controls, plus ownership, ties, oracle-field traps and invalid input.
- Full run: 9277 distortion identities, 10752 regret bounds; all ten frozen
  policies reproduce their previous A/B risks. All84 cells and both phases stay.
- Independent source audit: 258048 loss values, 32256 minimum comparisons,
  9277 Jensen identities, 8064 group means and 336 paired contrasts pass.
- Replay is byte-identical. Input, protocol and script hashes are recorded.
- No new inference-cost or serving-latency claim; this reuses stored forecasts.

Artifacts are `mmm-query-frozen-mass-v1.json`, `-replay.json`, and `-audit.json`.
The source-based audit explicitly marks the post-result Jensen investigation.

## Next Research Step

Audit the predicted joint dependence between a queried label and future labels,
not merely its marginal answer calibration. Use a coherent generator-level
joint model or held-out forward outcomes across independent trajectories;
evaluate conditional-response calibration with teacher identity kept entirely
in the evaluator. This targets the predictive-correlation lead already recorded
in the coverage literature. Do not interpret a Gaussian predictive-correlation
method as directly applicable to these Bernoulli mixtures without deriving and
testing the adaptation. All seven goals remain open; no production promotion.
