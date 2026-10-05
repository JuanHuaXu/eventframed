# Two-query observation lookahead: no robust rescue

Under the [frozen protocol](INTERVAL_LOOKAHEAD_PROTOCOL.md),1536 fresh-seed
episodes compare receding two-query target-risk lookahead with one-step target,
random, report-entropy and fixed policies. Every arm receives the same initial
four reports and purchases six renewals, total16 evidence credits.

Overall20/108 gates pass (FAIL):6/42 versus random,6/42 versus entropy, and8/24
final-loss protection checks versus one-step. The original84 comparisons
account for12 passes; the additional24 one-step checks account for8.
Do not compare20/108 directly to the earlier19/84: the seed set and gate count
differ.

## Pooled descriptive results on this fresh run

| Policy | Final Brier | Learning-area Brier | Accuracy | Confidently wrong |
| --- | ---: | ---: | ---: | ---: |
| Two-query | .370573 | .430565 | 73.96% | 30/1536 |
| One-query | .373239 | .430661 | 73.57% | 30/1536 |
| Random | .424402 | .458249 | 69.79% | 2/1536 |
| Report entropy | .375391 | .431883 | 73.44% | 11/1536 |
| Fixed | .420955 | .460453 | 69.79% | 0/1536 |

Two-query versus one-query pooled final Brier improves .002667, while
learning-area gain is only .000096. Area is the post-initial12-credit measure
defined in the protocol. These pooled values do not replace per-cell gates or
establish uniform superiority. Cross-run changes in absolute Brier cannot be
attributed to the policy because this run uses new seeds.

Of88 failed gates,18 also fail their point-mean threshold (eight final nonharm,
ten area gains);70 have acceptable means but inadequate descriptive lower
bounds. The evidence therefore does not support explaining failure solely by
sampling uncertainty. Confidence counts are retained, but do not on their own
prove or disprove calibration.

## Verification and information boundary

Full1536-episode replay is byte-exact. All7680 budgets,15360 score/area
recalculations and21344 copied-root outcome checks pass. Sampled histories
support1152 deterministic action reconstructions and1440 forecast-prefix
reconstructions using observed history only.

Independent depth2 branch enumeration passes24 cases, and24 depth1 checks match
the old one-step policy. Terminal horizon and history immutability checks pass.
The existing448 direct-quadrature forecast and192 cached-choice checks also pass.

The second planned action may depend on a HYPOTHETICAL first outcome; real
selection re-plans after the purchased observation. It never receives actual
future measurements. At depth2, up to72 child histories are evaluated, compared
with8 for one step. Equal evidence credits do not mean equal planner CPU, and
this run is not a throughput benchmark.

[Full traces and gates](interval-lookahead-experiment.json),
[verification](interval-lookahead-verification.json),
[planner](interval-lookahead.mjs).

## Next diagnostic

Test the full six-query Bayesian planning optimum within this finite model,
rather than indefinitely adding receding lookahead depths. The declared
renewal likelihood is exchangeable within each source: initial root outcomes
plus per-source renewal zero/one counts may be sufficient for belief and
remaining-budget planning. Verify that property before using it to compress
the planning state. An exact small-model reference could distinguish limited
planning depth from prior/model inadequacy.

This is a prospective lead, not a proof that full planning will help or scale.
The earlier impossible fixed-observation screen remains failed. Unknown
hypotheses, source authenticity, actual-agent tasks, population coverage and
loaded serving remain open. No production changes, whitepaper promotion,
performance claim or publication. All seven research directions remain open.

