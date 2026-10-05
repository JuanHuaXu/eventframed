# Returning-pattern V38 results

2026-10-03. **Frozen adoption FAIL30/48 design and37/48 confirmation cells.**
Late-shift gains pass the full component screen across both geometries and
both delay schedules in BOTH splits. Recurring score/recovery gains survive,
but some protection tests fail. This is not a general replacement, valid AP
authority or a completed whole goal. Production and whitepaper untouched.

## What Was Tested

The isolated OrientationObserver learns the first four genuine trial outcomes
of EACH member, then freezes their finite-shape posterior-predictive means.
Only after every anchor outcome arrives does a two-state HMM track whether
this template or its complement currently applies. Hazard1/600 is frozen.
This is exact filtering CONDITIONAL on a plug-in template, not ordinary
Bayes integrating uncertain member rates. An anchor-only control isolates
freezing from dynamic inference. Full64 and the V37 adaptive window bank
remain controls, without modifying their sources or prior negative evidence.

Delayed feedback is inserted at its original nomination position; the affected
forward suffix is replayed. Unobserved/cancelled positions have unit emission.
The ORIGINAL private emitted law is scored once. Nomination, not arrival or
wall time, advances the hidden state. Missing/cancelled anchor outcomes prevent
template publication. Epoch replacement is external, not a learned detector.

Fresh seed bases2026103803/2026103804 produce768 worlds,6144 arms,101376
snapshots and1843200 distinct conditional Bernoulli trials. Eight copies of
each world's2400 trials are not independent evidence. Two geometries and12
regimes include partial, unrelated, asynchronous and pre-anchor changes,
in addition to stationary/curved/abrupt/late/recurring/gradual controls.
Both immediate and150-tick delayed labels are independent of usefulness.
Outcome-dependent delay/noise negatives from V36 are NOT repaired here.

[Protocol](mmm-orientation-v38-protocol.md),
[preflight and research credit](mmm-orientation-v38-preflight.md),
[design audit](mmm-orientation-v38-design-audit.json),
[confirmation audit](mmm-orientation-v38-confirmation-audit.json),
[benchmarks](mmm-orientation-v38-benchmarks.txt).

## Gains and Limits

Wide-geometry confirmation; expected issued Brier, lower is better:

| Regime/schedule | Full | Adaptive window | Orientation | Whole cell |
| --- | ---: | ---: | ---: | --- |
| Late/immediate | .242325 | .209090 | .193701 | PASS |
| Late/fixed150 | .262063 | .237034 | .219157 | PASS |
| Recurring/immediate | .250738 | .248074 | .194766 | FAIL protection |
| Recurring/fixed150 | .269362 | .273223 | .243733 | FAIL protection |
| Abrupt/immediate | .251427 | .213901 | .192579 | FAIL protection |
| Partial/immediate | .230023 | .207911 | .246255 | FAIL, actual harm |
| Unrelated/immediate | .229595 | .208409 | .246870 | FAIL, actual harm |
| Early/immediate | .220700 | .204484 | .250884 | FAIL, actual harm |
| Curved/immediate | .196079 | .196123 | .205700 | FAIL protection/harm |

Late immediate whole gain .048624 has paired interval[.044265,.052983],
priority gain .048053 interval[.044168,.051938]. Restricted recovery is
5->2 rounds versus both full/window controls. The5 is the missed-deadline
penalty, NOT observed full/window recovery. With150-tick delay, confirmation
both geometries recover in3.5 rounds versus penalty5; wide whole gain .042905
interval[.039592,.046218], recovery gain1.5 interval[.717376,2.282624]. All
four late cells pass BOTH cohorts' quality/protection/cost checks.

Recurring immediate wide whole gain .055972 interval[.052583,.059361] and
priority gain .052781 interval[.049106,.056456] pass. Recovery penalty5 falls
to2.125 rounds, improvement2.875 interval[2.4375,3.3125]. With delay, issued
Brier improves to .243733, reversing V37's actual delayed-recurring harm;
recovery is3.104167. These risk/recovery components survive both geometries
and splits. They do not make the full recurring cells pass: confirmation
wide final usefulness protection versus adaptive has lower-.019483 immediate
and-.01875 delayed. These intervals span zero and are INCONCLUSIVE protection,
not proof of population usefulness harm. Tight immediate final risk/priority
protection also fails; tight delayed passes confirmation but not design.

Abrupt gains/recovery also survive, but confirmation final protection versus
adaptive fails. Wide immediate usefulness gain-.0075 interval[-.03375,.01875]
does not establish noninferiority. Do not select only the successful metrics.
Wide asynchronous passes both splits/schedules; tight confirmation fails
protection. There is no declared discrete-global recovery metric for this case.

Partial/unrelated are actual issued-score harm, not merely low power. Wide
partial gain versus full is-.016232 interval[-.019043,-.013421]; unrelated
gain-.017276 interval[-.020788,-.013763]. Expected packet usefulness falls to
.50375/.53375 versus adaptive .8. Early changes contaminate the four-label
anchor; its later law stays near .25 and recovery receives penalty15.
Curved stationary packet usefulness falls .871694->.774843 versus adaptive,
gain-.096851 interval[-.135771,-.057931], despite no regime change.

Intervals are mean +/-3.5SE across16 independent worlds/cell, not confidence
sequences, simultaneous AP coverage or field-population guarantees. Whole
adoption needs all48 cells in BOTH splits; observed failures are retained.

## Post-Hoc Representational Floor

[Diagnostic](mmm-orientation-v38-floor.json) and
[independent Node script](../../research/orientation-v38-floor.mjs) use future
true rates ONLY after collection/auditing. They are never model inputs, new
confirmation evidence, threshold changes or a proposed deployable oracle.

For a frozen template a, any global state-mixture law has
q_i=a_i+(1-2*a_i)*f,0<=f<=1. Weighted least squares gives its most optimistic
future Brier at f=clip(sum_i w_i*(1-2*a_i)*(p_i-a_i)/sum_i w_i*(1-2*a_i)^2).
This oracle even allows endpoints the positive-hazard filter cannot attain.
Seven direct floor/domain tests pass; both inputs match audited raw hashes.

Confirmation wide final oracle Brier:

| Regime | Oracle floor | Floor minus adaptive final Brier |
| --- | ---: | ---: |
| Partial | .249846 | +.073369 |
| Unrelated | .249329 | +.077268 |
| Early | .249857 | +.082173 |
| Curved | .198178 | +.019094 |
| Stationary independent | .184033 | +.022856 |
| Late | .182990 | +.000240 |

All16 wide partial/unrelated/early worlds in BOTH cohorts have oracle floor
>.20: faster filtering alone cannot meet the declared final Brier target for
these realized templates/rates. This is a representational limit, not a
population theorem. Curved confirmation oracle-minus-adaptive interval
[.017089,.021098] still exceeds .01; stationary independent interval
[.019611,.026100] does too. Freezing scarce member evidence loses continued
learning even with perfect orientation selection. Both the model family and
template-learning policy need improvement, not consumed-cohort retuning.

## Verification and Performance

Both full audits verify384 worlds/50688 snapshots, exact original-law and
receipt replay, separate batch Beta/atom anchor integrals, independent
two-entry matrix filtering, window-control batch laws, packet sorting/risk/
priority/recovery arithmetic, readiness, timing and domain coverage. Hidden
path enumeration independently checks short histories. The readiness boundary
is599 immediate/749 delayed in every world, determined by arrival of all600
anchor labels, not known change timing.

Twenty-three precollection source hashes, three postcollection auditor-source
hashes and benchmarks bind the record. Eleven tape mutations, changed-source
and missing-benchmark rejection pass; original receipts, ownership, replay,
normalization/freeze atomicity, cancellation, epoch and future-prefix tests
pass. Four research modules pass race tests(150.410s dispersion) and vet.
Go replay is not an independent-language generator; the floor arithmetic is
separate Node code. Candidate, gates and tapes stayed unchanged after collection.

Apple M4/Go1.27.1 isolated component:150-emission suffix replay .788-.791us,
1800-emission replay9.438-9.470us,150 filter forecasts .227-.229us, all zero
allocations. Replay benchmarks include suffix restoration and OBSERVED
emissions. Constructor155.413-155.883us/1418901-1418902 bytes; freeze plan
178.329-179.343us/379473 bytes. Their allocation sum1798375 bytes passes8MiB.
The freeze plan and construction are separate costs, not a single latency.

Maximum orientation2400-label phase16.220ms design/16.849ms confirmation,
including construction, anchor updates/freezing, issue, late resolve and
snapshots. All-arm maxima255.385/276.598ms also pass400ms. The adapter still
pays the reused ledger's otherwise unused378-kernel baseline dot product on
Issue; the sub-microsecond forecast benchmark measures the FILTER only.
Post-anchor filter prediction is O(1); suffix update O(affected history),
worst N*60 positions. Anchor/freezing are separately bounded. No cost of
acquisition, persistence, served requests or loaded freshness is included.
Design overlapped race tests and confirmation overlapped design auditing;
timings are conditional on that load, not isolated cohort speed comparisons.

Raw SHA256:

- design:`5e4d18541014a772b90ddc100e1da6bc25350d42dacd1c636f1c2d914b270e31`
- confirmation:`4cce51311afc3e5351773228fb5d0db16d5cf9b8d4b40538971bd510ca0632a5`

## Next Actions, Not Completed Goals

Keep this returning-pattern learner as a challenger, not a replacement.
Prospectively test a broader/falsifiable model with continuing template
learning, uncertainty integration and local changes; retain these negative
controls. A properly scored expert mixture/fallback is another lead, not a
guarantee that unsafe structural assumptions become AP-authorized groups.
Do not retune the anchor size/hazard or mix weights on consumed confirmation.

Prioritize a new untouched real-retrieval domain and loaded durable learning
alongside, rather than another stationary-only component. No valid faster AP
splitting, untouched agent benefit, durable loaded-service freshness or
equal-total-cost observation superiority follows from V38. All seven whole
goals remain OPEN; nothing is deployed, installed, committed or pushed.
