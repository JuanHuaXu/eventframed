# Closed-loop uncertain renewal: incomplete rescue

The [frozen protocol](ADAPTIVE_RENEWAL_PROTOCOL.md) was run on640 new synthetic
episodes, two splits, five cases and six policies. Every arm spends16 acquisition
credits. The uncertain model chooses its own observations in this experiment;
it does not inherit the old planner's trace. Screen: FAIL,10/32 gates pass.

## Confirmation split

| Case | Regular Brier | Certain-fresh mixed Brier | Uncertain mixed Brier | Uncertain renewal calls |
| --- | ---: | ---: | ---: | ---: |
| Independent20 | 0.203460 | 0.067013 | 0.147530 | 2.234 |
| Copied20 | 0.464500 | 0.219493 | 0.272410 | 3.172 |
| Mixed20 | 0.268843 | 0.298353 | 0.223989 | 2.484 |
| Copied05 | 0.097364 | 0.015431 | 0.037154 | 2.375 |
| False renewal20 | 0.378840 | 0.589333 | 0.378271 | 2.375 |

Uncertain mixed improves copied20 versus regular by0.192090
[0.016759,0.367421], and beats its matched random/entropy controls there.
But it fails protection against the certain-fresh planner on genuine renewals.
The split1 false-renewal gain versus certain-fresh is0.211062
[-0.034739,0.456864], which fails its positive-lower-bound criterion despite
a large mean gain. Its near-zero difference from regular also has an interval
crossing the-.01 protection threshold. Do not call those uncertainty failures
proof of mean harm, or discard them to obtain a passing screen.

Of22 failed gates, six also violate the point-mean nonharm ceiling: all six
compare with certain-fresh mixed on genuine renewal cases. Sixteen fail only
the interval requirement given their observed means. This is a post-run
diagnostic, not a revised test. Increasing sample size cannot be assumed to
repair the six mean regressions; their uncertainty still precludes asserting
population harm from this pilot alone.

Compared with the certain-fresh mixed planner, the candidate uses roughly2-3
renewals rather than5-6. It is not merely relabeling confidence; actions change.
On false renewal20, confidently wrong outcomes drop from19/64 to1/64, while
accuracy rises from70.3125% to71.875%, matching regular-only accuracy. That is
useful behavior, but it does not establish the full accuracy/protection claim.

Intervals are paired mean +/-3.3 SE over64 episodes, descriptive not simultaneous
or anytime-valid. Both splits and all32 gates remain in the artifact. Different
case samples must not be compared across v11/v12/v13 as paired improvements.

## Verification

Component tests establish branch ownership under copy-on-write, original-control
parity and action reconstruction from only the observed prefix. The previous
latent-model tests cover renewal-before-original and joint mode/root semantics.
The same prior and credit prices remain frozen; no model tuning follows results.

The [independent verifier](verify-adaptive-renewal.mjs) uses batch enumeration
of root/ordinary/fresh modes for each acquired history, with F fixed1 only for
the original controls. It reconstructs all56,410 forecasts across3,840 arms,
checks all final and credit-area scores, credit/slot accounting and source hashes.
Maximum forecast discrepancy is2.188e-14. This validates finite-model arithmetic,
not all real-world dependence or the quality of a two-credit planning horizon.
The full640-episode experiment was rerun: every action, outcome, forecast and
computed summary reproduced exactly. The [replay record](adaptive-renewal-v13-replay.json)
binds that check to the saved artifact hash.

Artifacts: [all traces](adaptive-renewal-v13.json),
[all scores and gates](adaptive-renewal-v13-summary.json),
[independent verification](adaptive-renewal-v13-reference.json).
The reference stores likelihood tables immutably and only copies changed branch
state. Planning computation is nevertheless separate from acquisition credits;
this study is not a loaded service benchmark or a claim of affordable LLM calls.

## Remaining question

The tested model learns freshness separately for every observation type. If a
single measurement mechanism governs multiple types, that can waste evidence
about the mechanism. A possible next model could represent shared versus
type-specific freshness explicitly and infer which is supported. It must also
include mixed genuine/counterfeit types as a negative control; blindly pooling
freshness would recreate the original trust failure. This is an untested lead,
not a proven explanation of these failures or a new authenticity certificate.

The earlier known-mode depth2 optimum does not prove two-credit optimality in
this larger unknown-mode model. A future diagnostic may also need to separate
model structure from the value of learning the source mechanism before reuse.
Neither possibility licenses relaxing old gates. All research directions remain
open; no production or whitepaper changes.
