# Repeated-trial dispersion V34: completed study

2026-10-03. The frozen overall component screen **FAILS**:6/24 design
and5/24 confirmation cells fail. Learnable dispersion repairs several
variance-mismatch components but does not solve all retrieval/calibration
requirements. All seven whole goals OPEN; production and whitepaper untouched.

## Experiment and Evidence

The [preflight](mmm-dispersion-v34-preflight.md) and
[protocol](mmm-dispersion-v34-protocol.md) precede fresh design2026103403
and confirmation2026103404. Across1,536 worlds, four learners receive
the same3,686,400 trial observations. Five forecast arms at five checkpoints
produce38,400 snapshots. Trials are independent CONDITIONALLY on each
member's persistent rate; they are not millions of independent trajectories.
Confidence intervals use32 independently generated worlds per cell.

Adaptive dispersion retains the27 affine mean functions and priors, with
an independent uniform prior over concentrations {.5,2,8,32,infinity}.
Controls are fixed2, point-mass/shared-only, independent local Beta2 and
static baseline. The infinity branch is phi=p, never Beta(0,0). Integrating
each persistent member rate gives the same ordinary likelihood and future
predictive law. No generalized-posterior weighting is introduced here.

- [Design raw tape](mmm-dispersion-v34-design.jsonl), SHA256
  `c1fc46e8d97c7411b40491f2aebd900da23840f294525a59c4a7a46c3ff179ab`.
- [Confirmation raw tape](mmm-dispersion-v34-confirmation.jsonl), SHA256
  `ab3c9921ba394f9d23021acac8dbb86840b468f7454961e89050a407058ccd87`.
- [Design audit/gates](mmm-dispersion-v34-design-audit.json) and
  [confirmation audit/gates](mmm-dispersion-v34-confirmation-audit.json)
  retain every checkpoint, control, interval, five frozen source hashes
  and the post-collection auditor's own hash.

## What Improved

At32 and150 labels the concentration weights remain exactly .2 each.
One distinct trial/member has Bernoulli(p) likelihood independent of
concentration. Different future predictions at those budgets come from
the newly declared PRIOR, not evidence that learned dispersion. The
frozen fixed2 control agrees with V27 through single-trial histories.

New trials of already observed members identify dispersion in the declared
model. Confirmation means at2400 labels (16 trials/member):

| Case | Fixed2 Brier / usefulness | Adaptive Brier / usefulness | Adaptive signed packed bias |
| --- | --- | --- | --- |
| Tight aligned | .205638 / .848674 | .195969 / .875839 | .000189 |
| Wide reversed | .205580 / .848003 | .195960 / .875772 | .000121 |
| Wide shared mean | .178928 / .973995 | .170661 / .976386 | .000792 |
| Wide Beta2 baseline | .132162 / .990529 | .132162 / .990529 | .002094 |
| Wide concentration .5 | .060087 / .999640 | .059239 / .999640 | -.000462 |
| Wide concentration8 | .160097 / .990561 | .158671 / .990561 | .000748 |
| Wide concentration32 | .174743 / .981286 | .169625 / .981140 | .001247 |

These are synthetic future expectations, NOT actual agent-answer accuracy.
Tight-aligned whole-Brier gain .009669 [.009032,.010306] and
wide-reversed .009620 [.009048,.010191] pass frozen .005/positive-lower
criteria in confirmation and design. Their priority-risk improvements
also pass. The strongest pooling branch receives about .956/.963 mean
posterior mass respectively in confirmation; shared-mean wide receives
.874. This addresses the fixed Beta2 overreaction when little true member
variance exists. It does not prove a universal cause of V33 failure.

Where the generator genuinely has Beta2 member variance, adaptive reverts
to fixed2 and preserves its forecasts. In wide concentration .5,8,32 cases
the correct component receives approximately1,.966,.982 mean mass at16
trials, respectively. At4 trials these latter components are less sharply
identified. This is ordinary model-conditional learning, not a coverage
certificate or authority to share real events.

## What Still Failed

Tight-calibrated whole-Brier gain .004807 confirmation and .004973 design
falls below the declared .005 mean target, despite positive lower endpoints.
Tight shared-mean gain .002587/.002684 also misses .005. These are real
improvements too small for that predeclared screen, not evidence of harm;
the thresholds are not changed to count them as passes.

Tight-curved usefulness protection against local fails in both splits:
confirmation mean -.002631, lower -.013792 below -.01. Wide-curved design
also fails (mean -.005260, lower -.013352), though confirmation passes.
An affine mean family remains mismatched to a curved truth; dispersion
learning becomes nearly identical to fixed2 in these cases.

Both independent two-rate geometries fail packed bias in both splits.
Wide-independent confirmation usefulness is .800000, but mean bias
.098816 implies mean packed forecast about .898816. Its magnitude upper
.105998 exceeds .10; tight-independent upper .100010 also just fails.
Design uppers .104007/.105623 fail. This is not merely an uncertain tiny
mean defect: the fitted continuous Beta member distribution does not
represent the generator's bounded two-point rate distribution. Knowing a
variance/concentration is not knowing its full shape. More trials helped
several risks but did not make the current family a general rescue.

## Audit and Performance

Both exact Go seed replays cover all generated rates, outcomes, nominees,
trial ordinals, original issued laws and25 snapshots/world, excluding
only explicitly elapsed fields. Separate batch log-Beta integrals verify
full joint weights and150 future laws at all checkpoints, using a different
calculation from incremental predictive-likelihood updates. The scorer is
checked with q^2-2qp+p, explicit150/170/10 denominators and stable packet
sorting. Near-ties sort by the exact recorded/replayed law rather than
tolerance-sized batch reconstruction noise.

This is an independent mathematical calculation within Go, NOT an
independent-language auditor. It shares the declared mean grid with the
model; single-trial comparison to the older independently coded V27 mean
family guards that boundary. All frozen source hashes are checked before
replay. Six corrupted-tape controls reject issued laws, outcomes, ordinals,
final laws, dispersion weights and phase sums; a seventh control rejects
an altered source hash without modifying source files.

Unarrived-trial flips leave past issued laws/snapshots unchanged. Full
batch checks through64 trials/member, one-trial nonidentifiability,
second-trial learning, replay/skips/cap rejection, ownership/nonmutation,
atomic invalid normalization and finite-log underflow recovery pass.
All four research-module race tests and vet pass. New post-collection
tests/auditor do not change the frozen candidate or either outcome tape.

The [Apple M4 benchmark](mmm-dispersion-v34-benchmarks.txt) measures150
forecasts13.818-13.889us, fresh updates1.469-1.475us and16th-trial updates
1.480-1.484us with zero allocations. Update timings include resetting the
saved benchmark state. Construction18.174-18.189us uses35712 bytes/four
allocations. Construction/later-update measurements ran alongside audit
jobs; these are component costs, not a serving or loaded-latency result.

Maximum measured per-model phase sum9.324929ms design/8.632581ms
confirmation passes25ms. Nomination is separately measured and shared;
world elapsed covers the four interleaved learners, logging and snapshots.
Expected-risk evaluation, acquisition, databases and queues are outside
learner costs. Static/local bookkeeping is bounded in the experiment;
reported phase sums are not a standalone complete wall benchmark.

## Next Research Boundary

The candidate is promising for repeated genuine outcomes and low-variance
means, not a universally superior challenger. Preserve its negative
gates. A next model-shape comparison should retain this hierarchy and
include a coherent bounded/discrete member-rate alternative on fresh
outcomes, rather than clipping forecasts to consumed true rates. Nonlinear
mean structure must also be evaluated separately.

Before integration, test delayed trials, regime shifts and stale snapshot
rejection; ordinary stationary hyperparameter inference does not supply
adaptation or stability. Actual agent evidence may be far rarer than this
fully labeled16-trial fixture. No AP authority, untouched agent improvement,
loaded freshness or equal-total-cost falsification win is established.
