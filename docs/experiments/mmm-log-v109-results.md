# v109 log-score advice: stronger partial result, overall FAIL

Both candidates were implemented and evaluated on all1,536 schedule-runs.
No candidate passes every required gate; neither is promoted.

| Candidate | Non-harm pass | Gain pass | Total pass | Verdict |
| --- | ---: | ---: | ---: | --- |
| Log/no-neutral | 478/480 | 28/38 | 506/518 | FAIL |
| Log/neutral | 470/480 | 28/38 | 498/518 | FAIL |

Every failed gate is in the delayed, late parity-to-majority case. Stationary
protection and all other case-specific gates pass in this finite screen. That
does not establish general stationary safety or close the roadmap.

The [frozen protocol](mmm-log-v109-protocol.md) preserves768 latent trajectories,
12 cases, two phase-disjoint rule pools,32 trajectories per cell and paired
immediate/delayed-missing schedules. Priors and gates are unchanged; the new
selector uses the Bernoulli log likelihood rather than eta=.5 Brier loss.
There is no age discount. Brier remains the primary evaluation criterion.

## Confirmation delayed late-half results

Expected binary Brier, lower is better:

| Case | Generic | Conservative | Arrival Brier | Brier-neutral | Log/no-neutral | Log/neutral |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Parity4 | .072770 | .071097 | .060804 | .060833 | .050526 | .050521 |
| Majority to parity | .254235 | .250397 | .251568 | .251635 | .214316 | .227004 |
| Parity to majority | .226968 | .226511 | .229176 | .229184 | .227765 | .238109 |

Log/no-neutral majority-to-parity gain versus arrival is .037252 with paired
interval[.026382,.048123]; versus conservative it is .036081 with
interval[.025133,.047029]. Both pass. Stationary parity4 expected accuracy
increases from94.7583% for arrival to94.9121% for either log candidate.
These are synthetic expected accuracies with5% outcome noise, not real-agent
answer rates or a reproduction of the original MMM accuracy experiment.

The reverse change remains unresolved. Log/no-neutral confirmation Brier harm
versus conservative is only .001255, but its interval[-.010117,.012627] fails
the .01 upper-harm allowance. Required gain also fails. Log/neutral is worse:
harm .011599, interval[.001117,.022081]. Its design result is adverse too.
Higher classification accuracy in this case does not rescue worse probability
quality: log/neutral accuracy70.0171% accompanies Brier .238109, versus arrival
accuracy69.1162% and Brier .229176.

Intervals retain the declared paired mean +/-3.5 standard errors. They are
approximate fixed-sample screens, not confidence sequences, exact coverage or
history-wide multiplicity control. All1,036 gates and secondary log scores
remain in the [machine-readable summary](mmm-log-v109-summary.json).

## Posthoc block diagnosis

The separate [block artifact](mmm-log-v109-blocks.json) covers both phases,
both switches and both schedules without filtering trajectories. It is
consumed, descriptive evidence, not new gates. The independent script checks
that all block scores reconstruct the original full and late means.

Confirmation delayed parity-to-majority Brier:

| Frame block | Arrival | Log/no-neutral | Log/neutral |
| --- | ---: | ---: | ---: |
| 128-159 | .391224 | .420398 | .406878 |
| 160-191 | .291285 | .266990 | .257930 |
| 192-223 | .164292 | .116871 | .188554 |
| 224-255 | .069903 | .106802 | .099076 |

The issue is not one uniform slowdown: there is early harm, mid-recovery
improvement, and late harm. Explicit neutrality reduces the initial harm but
substantially worsens192-223. It must not be prescribed simply because the
earlier hindsight oracle found some neutral headroom.

Next trace issued raw forecasts, selector and routed weights, observation masks,
loss origins, and model movement across publication. This will distinguish
stale role weights, changed fitted models and acquisition effects. The block
table alone does not prove which mechanism causes the regression. Preserve
the [handoff diagnosis proposal](../../research/model-handoff-diagnosis-proposal.md)
before implementing a new rescue.

## Verification and cost

- Literal nonuniform hidden-expert path enumeration matches the immediate
  ungated recurrence with share0/.001/.2, neutral enabled and disabled.
- Zero-share permutation, extreme log-weight recovery, strict-support and
  invalid-input checks pass. Log weights retain recoverable evidence when
  displayed probabilities temporarily underflow.
- Delayed single-use delivery, expiry, gate-version isolation, reentrancy and
  failed-acquisition atomicity pass. Component race tests:1.275s.
- All five controls match frozen v108 on consumed compatibility trajectories.
  Compatibility/seed race checks:6.352s; package vet passed.
- Generation199.91s; complete replay203.66s. All1,536 records match exactly.
  Independent JS reconstructs Brier/log scores, costs, as-of training origins,
  both update clocks and33 source hashes. Summary and block artifacts reproduce
  byte-identically. This is integrity evidence, not a delayed-martingale proof.

After quality, replay and verification stopped,256-frame fixed-model delay8
journal lifetimes measured1.609-1.640ms arrival,1.630-1.645ms Brier-neutral,
1.606-1.618ms log/no-neutral and1.608-1.622ms log/neutral. Log variants allocate
about151,481 bytes and1,241 objects per lifetime, versus about148,017 bytes
and1,219 objects for arrival. Policy-dependent acquisition prevents interpreting
these as pure arithmetic overhead or a significant speedup.

The complete seven-arm fixture, including model fits and records, took
130.05-145.91ms immediate and127.62-128.10ms delayed, about26.9MB allocated.
All repetitions are retained in [raw benchmarks](mmm-log-v109-benchmarks.txt).
No database, network, production serving or p95/p99 is measured.

Artifact305,299,146 bytes, created exclusively with mode0600. SHA256:
`6173c8d4c0232608876af25aaeba5e1ffbfddcf541c7c0870d51720ea03ca1ad`.
No production change, commit, push, private-data collection or whitepaper
promotion occurred. All seven research directions remain open.
