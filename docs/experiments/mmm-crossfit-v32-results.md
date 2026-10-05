# Exact cross-fitted combination V32: completed study

2026-10-03. Both frozen overall rescue screens **FAIL**: 20/24 design and
21/24 confirmation cells fail. Several forecast/retrieval components
improve, but broad calibration and protection do not pass. No whole
research goal is complete; production and the whitepaper remain untouched.

## Frozen Study and Artifacts

The [protocol](mmm-crossfit-v32-protocol.md) and
[preflight](mmm-crossfit-v32-preflight.md) precede collection. Two fresh
seed bases produce 1,536 worlds and 26,112 arms. All15 V31 controls are
retained on the NEW worlds; two exact-crossfit policies are added. Each
non-baseline arm gets32 distinct labels. Five primary stratum controls
receive exactly the same nominee/outcome/probability tapes.

- [Design raw tape](mmm-crossfit-v32-design.jsonl), SHA256
  `42cef291a05abd2eeff4be4ac39f4ad7cccb869418d6583871a58c6dcfa74e89`.
- [Confirmation raw tape](mmm-crossfit-v32-confirmation.jsonl), SHA256
  `a33bc8a0e98d06f3ed601b0e89d0c95368bf9c56b1de82d818cf8f0f38ffd9af`.
- [Machine-readable summary](mmm-crossfit-v32-summary.json) contains
  every gate, interval, metric and21 frozen source hashes.

For each arrived label, child validation excludes its OWN outcome from
both model weights and leaf counts. All other CURRENTLY arrived labels
may enter. This is current-prefix LOO fitting, not a historical issued
prequential score. Final laws combine full-data child predictives. The
same fixed simplex, SUM-loss ridge1 and (.98,.01,.01) anchor are retained.
Fitted predictive weights are not Bayesian family probabilities.

## Quality Results

Confirmation means for primary random-within-stratum arms:

| Case | Affine Brier / usefulness | Blend Brier / usefulness | Prequential stack Brier / usefulness | Crossfit Brier / usefulness |
| --- | --- | --- | --- | --- |
| Tight curved | .252460 / .630888 | .223875 / .724577 | .240293 / .735772 | .225030 / .752486 |
| Wide curved | .254404 / .597058 | .254611 / .493504 | .258646 / .493003 | .243634 / .548863 |
| Wide independent | .256112 / .666875 | .281896 / .573125 | .272550 / .601250 | .267027 / .614375 |
| Tight reversed | .212446 / .835705 | .218070 / .816745 | .239133 / .790218 | .233015 / .783876 |
| Wide reversed | .208016 / .844664 | .209585 / .831594 | .228253 / .722685 | .223089 / .738322 |

Lower Brier and higher expected ten-item packet usefulness are better.
These are synthetic model-world future expectations, not agent-answer
accuracy. Same-world paired comparisons are valid; differences between
V31 and V32 tables are NOT paired because their world seeds differ.

Tight-curved confirmation improvement versus affine passes all THREE
declared component thresholds: whole-Brier gain .027430
[.013803, .041058], priority-Brier .024947 [.008897, .040997], and
usefulness .121599 [.050310, .192887]. These are frozen mean+/-3.5SE
intervals across32 worlds, not confidence sequences. Design whole and
priority-Brier lower endpoints remain negative, so the two-split
improvement requirement is not established.

Even that confirmation cell fails blend priority-risk protection: lower
gain -.016035 falls below -.01. Its packed bias magnitude upper .135191
also exceeds .10. Thus the attractive component numbers cannot override
the overall failure.

Wide-curved whole and priority Brier gains versus blend pass in BOTH
splits. Confirmation gains are .010977 [.002663, .019291] and .014793
[.001890, .027696]. Usefulness gain .055359 has interval
[-.020617, .131334], failing the positive-lower-bound criterion. Packed
bias mean .279872 and magnitude upper .489799 show substantial confidence
overstatement that more trajectories alone cannot make disappear.

Wide-independent confirmation improves versus blend on all three requested
means/intervals, but design usefulness uncertainty fails and affine
protection fails. Confirmation packed bias upper .264025 fails .10.
Wide-reversed usefulness loss versus affine is -.106342
[-.160708, -.051976]; both forecast risks also fail affine protection.
The new construction is not a general rescue for reversed cases.

Only tight/calibrated, tight/baseline_matched and wide/calibrated pass all
confirmation cell gates. Design additionally passes wide/baseline_matched.
All failed cells, controls and original thresholds remain in the summary.

## Audit and Cost

Independent JavaScript reconstruction passes for all26,112 arms: integrated
child likelihoods, pre-outcome selected laws, final laws, explicit omitted
rows, fitted simplex weights using a different analytic solver, nominees,
packets, risks, timings and21 source hashes. The auditor's60 pre-collection
full-omission/own-label checks pass. Historical adaptive-control maxima are
not separately recomputed by JS here; exact Go seed replay and their
earlier unit/study audits cover those choices, as the protocol states.

Exact BOTH-split Go replay passes all non-timing fields, including generated
rates, latent draws, nominee RNG decisions, old issued rows and new final
LOO rows. Only five elapsed fields per arm are excluded. Seven deliberately
corrupted-tape controls reject changed LOO rows, selected outcomes,
historical laws, final laws, fit weights, packets and source hashes.
See [negative controls](mmm-crossfit-v32-negative-controls.json).

Full omitted-member refits and omitted-own-label flips agree within2e-13
for2,11,150 and200 members, including all200 arrived labels. Rebuilt
combination weights/laws agree within5e-12. Nonmutation, input ownership,
duplicate/invalid/adaptive-nomination rejection, partial-failure quarantine,
collector unused-label flips, matched evidence, three-package race and vet
all pass. No service or production integration is made.

The [Apple M4 microbenchmark](mmm-crossfit-v32-benchmarks.txt) records
32-label weight refit2.756-2.830us,64-label4.979-5.036us,150-label
11.042-11.110us, and150 forecasts9.553-9.569us, all zero allocations.
Construction costs50.505-51.912us and76248-76249 bytes/117 allocations.
These are component costs, not full update or daemon serving latency.

For frontier size N and L arrived members, each refit scans N flags and
L rows of H=54 child components, plus a fixed-size three-weight QP:
O(N+LH). Refitting after every arrival costs O(LN+L^2 H) cumulatively.
This is bounded by N<=200 here, not constant in evidence volume and not
V31's constant-size incremental original-row fit. Timing is measured at
N=150; N=200 is covered by mathematical/refit tests, not a separate timing
claim. The quality/cost tradeoff must survive later budgets and load tests.

Maximum isolated model-arm total is1.069917ms design and1.458333ms
confirmation. Crossfit final timing includes extracting32 diagnostic LOO
rows. Totals exclude packet scoring, acquisition, database, queues and
agent output. Fixed32 comparisons do NOT equate total observation cost.

## Interpretation and Next Test

Exact mature-child validation improves several components at small
computational cost, but the frozen overall rescue is rejected. Arithmetic
and own-label omission are not the remaining explanation for failure.
Ridge anchoring, one-label-per-member uncertainty, model mismatch and
packed-selection effects remain distinct plausible causes.

Next use a newly frozen label-budget learning-curve study with matched
nonadaptive evidence. Keep model/regularizer constants unchanged and vary
available arrived-label count; distinguish information scarcity from
persistent model/calibration failure before adding another model or
retuning a prior. Retain the no-future/full-refit controls. More independent
worlds can reduce interval width but cannot fix a large mean confidence
defect; more evidence per world must be tested, not presumed beneficial.

No valid Anti-Pigeon sharing authority, delayed recovery, changing-regime
stability, untouched agent improvement, loaded freshness, or equal-total-cost
falsification win follows. All seven whole goals remain OPEN.
