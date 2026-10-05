# Paired measurements V54: component gain, broad rescue failure

## Verdict

Independent paired measurements identify the declared independent noise in
this diagnostic and usually improve issued risk over the unpaired learner.
They do NOT pass the broad quality, recovery or equal-total-cost observation
screen. No adoption or reserved confirmation dispatch. All seven goals OPEN.

[Protocol](mmm-paired-v54-protocol.md),
[raw outputs](../../research/paired-v54-diagnostic/diagnostic.jsonl),
[independent readback](../../research/paired-v54-diagnostic/readback.json),
[source freeze](../../research/paired-v54-diagnostic/freeze.json),
[terminal commands](../../research/paired-v54-diagnostic/completed.json).

40 worlds:20 regimes x2 geometries,3 arrival schedules,7 arms=840 arms.
96000 DISTINCT underlying outcomes Y. A requested W2 measures the SAME Y as W1;
192000 secondary requests across the four policies are NOT new independent
underlying outcomes. Each policy has2400 primary+400 secondary requests per arm,
including missing responses. Full, Adaptive and no-pair have only2400: resource
ablations, not fair2800-request selection controls. Keep all20 regimes, including
deliberately correlated sources and two20%-missing-secondary cases.

Diagnostic base2026105407 consumed. Design2026105409/confirmation2026105411
UNEVALUATED; seed metadata uniqueness checks do not evaluate them. One world
per geometry/regime, shared across delays/policies: n1, not120 independent
replicates. No confidence intervals or validated population claim inferred.

## Quality

Equal-cell means across120 geometry/regime/schedule cells. Lower Brier is better;
usefulness is the terminal top10's expected clean outcome rate, NOT agent-answer
accuracy. Rates and hidden outcomes are evaluator-only inputs.

| Arm | Issued Brier | Priority Brier | Terminal Brier | Top10 Usefulness |
| --- | ---: | ---: | ---: | ---: |
| Full,2400 requests | .224699 | .224199 | .208152 | .727150 |
| Adaptive,2400 requests | .214748 | .214288 | .181599 | .815038 |
| No-pair,2400 requests | .218761 | .217644 | .189120 | .808179 |
| Random,2800 requests | .217843 | .216680 | .188353 | .811310 |
| Uncertainty,2800 requests | .217792 | .216636 | .188557 | .810810 |
| Information,2800 requests | .217772 | .216611 | .188205 | .809310 |
| Falsification,2800 requests | .217772 | .216618 | .188172 | .809810 |

Random/uncertainty/information/falsification improve issued loss over no-pair
by.000917464/.000968265/.000989060/.000988359. This includes extra measurement
information, not a same-sample-budget challenger win. Falsification wins96/
loses24 cells versus no-pair, but only67/52 versus random (one tie). Its mean
advantage over random is.000070895 and over uncertainty.000020094. Information
is slightly better on mean issued loss; no universal concentration-policy win.

All four miss the unchanged.01 Full-relative mean-gain gate: gains are only
.006856-.006928, even before the absent positive-confidence-bound requirement.
All four have17 cells exceeding.01 Adaptive harm. None passes broad protection.

Across60 declared-change cells, random/uncertainty improve recovery2 cells
against no-pair; information/falsification improve1. None worsens against
no-pair, but falsification STILL averages.716667 rounds slower than Adaptive;
mean per-cell relative Adaptive recovery gain is NEGATIVE12.5648%, not the
required positive10%. Top10 utility versus no-pair improves only3 falsification
cells and worsens6. CLEAN recurring cases gain.06 utility in all six cells,
but CLEAN aligned cases average-.027383. Preserve both, not just the rescue.

## Measurement Identification And Its Boundary

Post-run inspection, not tuning: all clean regimes' paired falsification runs
place approximately all terminal noise mass on eta0. Independent noise10/20
cases place approximately all on the matching declared eta, including missing
secondary sources. This fixes V53's specific clean-change/noise-confounding
symptom in these generated cases; it does not supply authenticated truth.

Correlated round-noise sources agree with each other even when both are wrong.
The learner therefore infers eta0 there, and issued loss regresses by.001139564
against no-pair on average. Source independence is a load-bearing assumption,
not proven by agreement or by a large posterior. Partial_noise20 also loses
.000539378 against random and.000920400 against uncertainty. Successful noise
identification alone does not imply useful acquisition or fast recovery.

## Cost

Maximum complete2400-nomination/drain loop, not per-request latency:

| Arm | Maximum ms | Cells Above400ms | Total Proposal ms,120 Arms |
| --- | ---: | ---: | ---: |
| Full | 64.917 | 0 | 0 |
| Adaptive | 213.628 | 0 | 0 |
| No-pair | 176.898 | 0 | 0 |
| Random | 230.860 | 0 | 1.695 |
| Uncertainty | 424.924 | 22 | 20042.363 |
| Information | 415.931 | 28 | 20498.200 |
| Falsification | 413.868 | 25 | 20216.424 |

Falsification total elapsed45.779s versus random25.654s and uncertainty45.545s:
78.45% more than random and.51% more than uncertainty at the SAME2800-request
budget. No computational dominance or equal-total-cost superiority. Random is
not artificially charged unused scoring. Proposal, request, replay, snapshot,
scheduling and component totals are recorded; raw JSON serialization/auditing
is separate from the component loop and included in command wall time.

Constructor max3411344 allocated bytes passes8MiB, NOT a resident-memory bound.
Serialized fixed-order offline costs and finite n1 results do not prove loaded
100ms serving, freshness, production throughput or Goal6.

## Audit

Seven prospectively frozen commands terminate0: model race, fixture race,
closure helper, vet, allocation, full collection and full independent audit.
The independent audit reconstructs ALL40 worlds/840 arms, original-position
joint/marginal predictions, choice scores/batches, receipt order, missingness,
posterior weights, snapshots, issued losses and recovery. Separate readback
verifies source closure and recomputes aggregate metrics and costs.

Six model test roots cover lifecycle/atomic failure, caps/cancellation, contract
rejection, explicit paths, delayed joint reference and unobserved-evidence forks.
Two21^3 latent-path cases and256 first+256 second reference updates pass.
15 nonvacuous future forks,32 policy artifact corruptions and6 control aggregate
corruptions pass.3840 declared world seeds/23040 domain-separated channels are
distinct. This is the declared six-base namespace check, not all historical RNGs.

Before collection,60 actual repository compiler/test inputs+11 support files
were frozen/copied. Go1.27.1 darwin/arm64, compile/link hashes, host metadata and
logs saved. All14 pre-existing tracked edits retain their preceding checkpoint
hashes. [Preflight repairs](../../research/paired-v54-preflight.md) preserve
the additive RNG alias and the control-audit early-return regression; no
outcomes, model parameters, cohorts or success thresholds were retuned.

Production, private corpora, whitepaper and existing confirmation artifacts
remain untouched. No commit, push, install, deployment or default-store access.

## Next Action

Preserve this broad FAIL and do not consume its reserved confirmation cohort.
An exact computational rescue is worth testing, but cannot itself fix the
negative quality/recovery cells. Original-factor forward/backward smoothing
and member-revision-bound query caches are a separate prospective lead:
[direction and algebra checks](../../research/paired-v55-direction.md).
Future acquisition also needs explicit total-cost matching and source-dependence
coverage, not duplicated testimony treated as independent evidence. Untouched
agent tasks, useful split integration and loaded freshness remain separate work.
