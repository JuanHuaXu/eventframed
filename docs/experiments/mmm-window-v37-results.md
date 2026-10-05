# Rolling/adaptive windows V37: results

2026-10-03. **Overall frozen adoption component FAIL:8/32 design and12/32
confirmation cells.** Abrupt and gradual improvements survive both splits;
late/recurrent recovery and some stationary protection remain unresolved.
All seven whole goals OPEN. No production or whitepaper change or push.

## Implementation and Verification

The new isolated RollingShape retains the newest W arrived ISSUE ORDINALS,
not newest network arrivals. It removes an old outcome using its conditional
likelihood under the remaining member evidence, then adds a new outcome
under that SAME remaining set. This is an exact truncated-evidence working
posterior, not ordinary Bayes over full history or a correct latent-drift model.
Windows4/8/16/full64 remain as independent controls.

Adaptive mode retains all four children and aggregates their current laws.
Weights track their ORIGINAL issued log losses with frozen share1/600 and
restart prior(.85,.05,.05,.05). They are predictive expert weights, not
ontological posteriors or AP authority. Old late labels cannot evict newer
child evidence, but their original issued losses still update expert weights.
This arrival-order meta-update can reflect stale performance. It is not
ADWIN and inherits no detector or no-delay regret guarantee.

Fresh seeds2026103703/2026103704 give512 worlds,5120 arms,84480 snapshots
and1228800 distinct trials. Ten learner/schedule copies are not independent
observations. Two geometries/eight regimes/16 worlds per cell cover stationary,
curved/Beta-matched, abrupt, late, recurring and gradual cases, with immediate
and150-tick delayed feedback. No noise/selection rescue is presumed from this
clean-label study; V36's negatives remain.

Both full audits replay populations, every original issued law/expert row,
receipt and metric. Batch Beta integrals/direct two-atom products reconstruct
each child from its newest arrived identities, not its stored counts/masks.
Independent fixed-share arithmetic reconstructs every issue mixture and
snapshot; independent phase scans verify restricted recovery. This arithmetic
is within Go; RNG replay uses frozen generator code, not an independent
language implementation.

Eighteen precollection source hashes, separately hashed auditor/benchmarks,
ten corrupted-tape controls plus changed-source rejection, future-prefix
flips and owner/time/replay/cancel/epoch/all-child atomicity tests pass.
Four research modules pass race tests and vet. Model, tests, protocol and
cohorts were not retuned after collection.

[Protocol](mmm-window-v37-protocol.md), [preflight](mmm-window-v37-preflight.md),
[design audit](mmm-window-v37-design-audit.json),
[confirmation audit](mmm-window-v37-confirmation-audit.json),
[benchmarks](mmm-window-v37-benchmarks.txt). Raw JSONL files accompany them.

Raw SHA256:

- design:`933ecbec5229d04140d8f38dafde684f8b02b58fa47e4608439b1e01ed7a01c6`
- confirmation:`4aee0eec37d75a4b3d8c441f14bee9945a11bb83678666d880f2c12bbb5dc56a`

## Gains That Survive Confirmation

Wide-geometry confirmation means; smaller Brier is better:

| Regime/schedule | Full issued Brier | Adaptive issued Brier | Full final Brier | Adaptive final Brier |
| --- | ---: | ---: | ---: | ---: |
| Abrupt/immediate | .248844 | .213020 | .250000 | .179421 |
| Abrupt/fixed150 | .272548 | .240776 | .250000 | .185480 |
| Late/immediate | .243062 | .209367 | .326813 | .183699 |
| Gradual/immediate | .251025 | .227859 | .250095 | .184760 |
| Recurring/fixed150 | .269018 | .272465 | .250001 | .247679 |

Abrupt immediate issued gain .035824 has interval[.033207,.038441]; priority
gain .035570 has interval[.033049,.038091]. Restricted mean recovery falls
9->5.1875 completed rounds, with paired improvement3.8125 interval
[3.336583,4.288417]. The full-history9 is the predeclared missed-deadline
penalty, NOT an observed successful recovery. With fixed150 delay, recovery
falls9->6.1875; issued gain .031772 interval[.029218,.034326]. These pass
the primary abrupt criteria across both geometries and BOTH splits.

Gradual immediate issued gain .023167 interval[.020398,.025936] and priority
gain .023291 interval[.020397,.026185] pass. Both delayed gradual cells also
pass, as do the design counterparts. No discrete recovery-time criterion is
declared for gradual drift. Final wide top10 usefulness reaches .8 for these
confirmation abrupt/gradual cases.

Intervals are paired mean +/-3.5SE across16 independent worlds/cell, not
confidence sequences or simultaneous AP guarantees. Fitted samples and
regimes are evaluator definitions, not evidence of causal identification.

## Preserved Failures

Both cohorts fail the four late and four recurring geometry/schedule cells:

- Late: issued/final quality improves substantially, but neither model meets
  TWO consecutive qualifying round-end snapshots within the four-round
  changed phase. Both receive recovery penalty5. One good last forecast
  cannot retrospectively satisfy sustained recovery. No metric was loosened.
- Recurring immediate: confirmation wide whole gain .002636 and priority
  gain .000783 are positive but below the required .01. Recovery gain is0.
- Recurring fixed150: confirmation wide whole gain is NEGATIVE -.003447,
  interval[-.005138,-.001757], and priority gain -.003788,
  interval[-.005607,-.001968]. This is actual issued-score harm, not merely
  inadequate power. Final usefulness .785 conceals a poor learning path;
  final Brier .247679 remains close to uninformative .25.

Confirmation additionally fails four curved stationary cells on packet
usefulness protection. Tight mean gain -.002813 has lower-.012765;
wide mean gain -.005221 has lower-.018471, outside the -.01 tolerance.
Those intervals also span zero: they do not prove large population harm,
but they do NOT establish the required noninferiority. Design passed these
cells. Other stationary metrics/cells pass; do not summarize that as full
stationary protection across both cohorts.

Thus neither selecting only successful regimes nor quoting final accuracy
would establish the full requested recovery/protection goal.

## Performance and Limits

Apple M4 Go1.27.1: replacement18.176-18.753us, adaptive Issue1.141-1.149us,
first Resolve43.494-43.937us. Warm Resolve after16 rounds/member, with short
windows replacing old labels,75.585-76.072us. All allocate zero. Benchmarks
include state restoration. Construction347.770-349.648us and3071454-3071460
bytes/55 allocations. Predict/update are O(J*378), J<=4; storage depends on
N, child count and64 identity slots/member. Trials are bounded per epoch;
this is not indefinite autonomous continuous learning.

Maximum measured complete-arm learner phase sum217.434ms design/258.224ms
confirmation, below the predeclared400ms background component cap. This is
2400-label total work, NOT one request latency. Confirmation collection
overlapped the design audit, so timings are conditional on that local load,
not an isolated-CPU speed comparison between cohorts. Schedule building is
measured separately; acquisition, persistence and loaded serving/freshness
are untested here. Constructor/phase caps pass, quality adoption does not.

## Next Research Leads

Separate at least three possible causes of the short-phase failure: not
enough fresh per-member evidence, inappropriate latent-drift assumptions,
and original-loss expert adaptation that arrives too late. Retain this bank
and full-history controls. Prospective leads include explicit transition/
changepoint hypotheses that reuse a justified old pattern, predictable
age-aware expert credit, and held-out control of nonlinear packet churn.
These need new protocols, off-family/partial/asynchronous changes and fresh
outcomes; no confirmation-data tuning or source/AP self-certification.

Keep working on all seven goals: this partial component does not demonstrate
valid faster AP splitting, untouched agent benefit, durable loaded freshness,
or falsification sampling superiority at equal total cost.
