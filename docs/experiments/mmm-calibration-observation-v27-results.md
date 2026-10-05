# Calibration observation v27: broad rescue fails

2026-10-02. Both independent splits FAIL the frozen challenger and modeled-cost
observation screens. Keep the new model isolated. No production or whitepaper
changes, and no whole research goal is complete.

## Evidence and Method

[Protocol](mmm-calibration-observation-v27-protocol.md) was frozen before
collecting either cohort. Each split has384 worlds: two cosine geometries,
six regimes and32 paired worlds per case, with21 arms per world. There are
8064 recorded arm runs per split,16128 combined. Every policy uses the same
potential-label tape within a world. The model sees only selected labels.

- [Design](mmm-calibration-observation-v27-design.jsonl), SHA256
  `72bb240dd64a18d980592eb4ed251ae1e64eeeb69b73873fd9a68883c5c466c4`.
- [Confirmation](mmm-calibration-observation-v27-confirmation.jsonl), SHA256
  `f78b8c4b61d8452be0e2b93fb724dc705e54a15c681258cbf09d5d83e2d32688`.
- [Independent verifier](../../research/calibration-observation-v27-verify.mjs)
  checks five frozen source hashes, cosine baselines, regime rate formulas,
  the evidence tape, ordinary posterior likelihoods, issued/final probabilities,
  deterministic selector choices, information/uncertainty maximality, budget
  admission, packets, expected risks and all frozen gates.
- [Summary](mmm-calibration-observation-v27-summary.json) retains every failed
  check and paired mean+/-3.5SE interval. These are descriptive finite-world
  intervals, not simultaneous or sequential certificates.

The ordinary hierarchical model has27 global calibration hypotheses and
separate Beta-distributed event rates. Evidence and prediction are marginals
of the same joint family. The prior predictive equals the original baseline;
one observation per event is supported. A separate local Beta head32 control
retains the v26 type of per-event update without a global calibration model.
These are offline laws, not a new Service or OpenClaw execution.

## Fixed32-label Confirmation

Lower Brier is better; packet usefulness is the mean true outcome probability
of the selected ten events. Bias is mean forecast minus true packet usefulness.

| Geometry/regime | Local Brier -> information Brier | Local usefulness -> information usefulness | Local bias -> information bias |
| --- | ---: | ---: | ---: |
| Tight independent | .37959 -> .25186 | .67437 -> .69687 | .27530 -> .02327 |
| Wide independent | .28441 -> .25123 | .68000 -> .67437 | .26419 -> .01969 |
| Tight reversed | .35795 -> .20309 | .22918 -> .86544 | .70857 -> .03263 |
| Wide reversed | .35182 -> .20209 | .23867 -> .86581 | .65871 -> .02588 |
| Tight aligned | .39802 -> .20523 | .87173 -> .87117 | .07819 -> -.01027 |
| Wide aligned | .21566 -> .21454 | .87216 -> .87218 | .07659 -> .06674 |
| Tight calibrated | .09382 -> .09529 | .92490 -> .92465 | .02503 -> .04820 |
| Wide calibrated | .18923 -> .19044 | .92341 -> .92341 | .02553 -> .04675 |
| Tight curved off-model | .34807 -> .35380 | .31114 -> .23176 | .62812 -> .12212 |
| Wide curved off-model | .25539 -> .35978 | .32012 -> .22286 | .58397 -> .11014 |

Information acquisition passes the independent/reversed risk gains and
aligned/calibrated protection comparisons in both splits. It nevertheless
FAILS the packed-bias ceiling for curved patterns in both geometries/splits:
confirmation magnitude upper .19782/.16994 exceeds .10. Wide aligned
confirmation also narrowly exceeds the bias ceiling at .10243. This is a
real generalization limitation, not a successful calibration rescue.

Stratification handles the curved pattern much better in mean, but FAILS
aligned packet-quality protection, calibrated wide packet protection, and
several uncertainty-aware bias bounds. Tight aligned confirmation mean
usefulness loss .02411 has harm upper .03645, above .01. Do not select a
winner based solely on its best regime.

## Explicit Modeled-cost Comparison

At label cost100000, information uses32 observations, uncertainty35 and
random40 under the same cap. Compute savings buy the cheaper policies more
labels. There is no equal-label shortcut in these comparisons.

| Confirmation case | Brier: random / uncertainty / information | Usefulness: random / uncertainty / information |
| --- | ---: | ---: |
| Tight reversed | .20741 / .37937 / .20309 | .84703 / .39742 / .86544 |
| Wide reversed | .20858 / .31730 / .20209 | .83279 / .48413 / .86581 |
| Tight curved | .25338 / .35802 / .35380 | .64516 / .23867 / .23176 |
| Wide curved | .25668 / .31830 / .35978 | .54330 / .11006 / .22286 |
| Tight matched model | .06494 / .06403 / .06526 | .96489 / .96425 / .97469 |
| Wide matched model | .15873 / .15718 / .16531 | .94817 / .93358 / .95183 |

The Goal7 screen FAILS in both splits. Information about the global hypothesis
is not the same objective as future-event risk reduction or useful top-ten
selection. Even under the matched model, learning the global hypothesis can
be less valuable than observing more individual event outcomes.

Cost units are frozen accounting assumptions, NOT a proved conservative bound
on primitive operations or hardware time. In particular, bit reversal has
per-index bit work not represented separately by its modeled256-unit charge.
The source-of-label costs are hypothetical, not agent acquisition timings;
actual equal-total-cost observation remains an open goal.

## Audit and Performance

Two-package race tests, normal unit tests, `go vet` and `git diff --check`
pass. Joint enumeration checks the evidence/kernel relationship; negative
controls reject duplicate/out-of-range evidence without state mutation,
exhausted frontiers, missing RNG and invalid priors. A future-tape test holds
observed labels fixed while changing unobserved labels and obtains identical
forecasts. Independent verification reconstructs adaptive decisions using
only prior state and previously selected labels. The verifier does not
reproduce Go's random-index or Gamma RNG streams; their implementation is
source-hashed, and sampler mean/variance controls cover small/large shapes.

Three100ms benchmark repetitions on Apple M4:

| Operation | ns/op range | Allocations |
| --- | ---: | ---: |
| Head selection, initial frontier | 1.878-1.934 | 0 |
| Stratified selection, initial frontier | 6.458-7.118 | 0 |
| Uncertainty selection,150 events | 4528-4591 | 0 |
| Information selection,150 events | 7887-7926 | 0 |
| Construct27-hypothesis model | 46469-46615 | 67488-67490 bytes,6 allocs |

Initial head/stratified selection returns the first eligible index immediately;
those tiny numbers do not represent late-frontier scanning. Maximum recorded
model-arm elapsed time was .56067ms design and .77425ms confirmation, excluding
data generation, true-law scoring, packet sorting, database, queues and full
serving. Timings are elapsed call measurements, not CPU-process attribution.

## Next Falsifiable Lead

Test observation by expected reduction in future-event Brier risk, rather than
information about the global hypothesis alone. Include individual-rate
uncertainty and priority weights, charge its additional computation, and
compare against random, uncertainty and information on fresh worlds. A
nonlinear model remains a separate possible rescue; adding a known curved
feature after seeing this failure must be declared and tested against new
off-family controls. Do not tune this protocol on its consumed cohorts.

Reproduce: `node research/calibration-observation-v27-verify.mjs`.
All seven whole goals remain OPEN.
