# Declared-Mean And Calibration Breadth V41 Prospective Protocol

2026-10-03. Isolated research, frozen before diagnostic or normal outcomes.
All V40 candidates failed broad quality. A confirmed interpretation confound
is that density-grid strength changes actual first moments, not just shape.
This is NOT proven to cause the failures. Narrow calibration means, static
global calibration, rate-shape misspecification and noisy feedback remain
competing explanations. No upstream production bug or patch is asserted.
Use new packages and files; preserve V39/V40 code, failures and cohorts.

## Research And Patch Reasoning

[Wainwright and Jordan, Graphical Models, Exponential Families, and Variational
Inference, section 3.5](https://people.eecs.berkeley.edu/~jordan/sail/readings/wainwright-jordan-fnt.pdf)
gives the log-partition mean/covariance identities and strict convexity for
minimal families. Here the nonconstant statistic is a finite rate atom. We
use those identities to solve a declared first-moment constraint; no published
adaptive recovery, borrowing or runtime guarantee is inherited.

Let a_z=(z+1/2)/21, z=0..20. For declared mean mu strictly within the atom
range and strength s, let positive reference mass nu_z be the normalized
Beta-shaped density-grid weights a_z^(s*mu-1)(1-a_z)^(s*(1-mu)-1).
The moment alternative minimizes KL(p||nu) subject to sum p=1, sum p*a=mu:

```math
p_lambda(z) = nu_z exp(lambda*a_z) / sum_u nu_u exp(lambda*a_u)
A'(lambda) = E_lambda[a],  A''(lambda) = Var_lambda[a] > 0.
```

Unique finite lambda exists for an interior target. Fixed 80-step bisection
within [-4096,4096] must bracket and meet 2e-12 mean/sum tolerance or reject.
An independent safeguarded Newton solver and KL stationarity/feasible
perturbation tests falsify the implementation. Neither solver sees outcomes.
Matching means across s does NOT match higher moments or isolate variance
alone. The density alternative deliberately preserves V40's confound.

NARROW: mu_i=(b_i,1-b_i,.5), prior family weights (.8,.1,.1).
RICH: retain those three at weights (.72,.09,.09), add 25 affine means
clip(.1+.2*j+(-.8+.4*k)*i/(M-1),.025,.975), j,k=0..4, each weight .004.
The clip lies strictly inside the 21-atom convex hull, unlike the old .02/.98
clip. Duplicate predictive families are allowed; no identifiability claim.
Breadth changes initial prior mixtures; it is NOT a matched-initial-law
contrast. No inverse-center factor, because its largest mean exceeds this
grid's support. Rate-shape breadth and switching global families stay separate
future hypotheses, not silently included in this ablation.

All experimental arms share one static family. Conditional member rates
reset independently to the family prior with hazard 1/16 per nomination.
Joint outcome/evidence law and original as-of forecast scoring are as in V40.
Pending/cancelled emissions are unit likelihoods but consume transitions;
fixed/uniform evidence delays are ignorable, not arbitrary production MNAR.
Tickets bind original private forecast, owner, epoch, ordinal and time.
One owner; explicit epoch reset, no automatic detector/AP/graph authority.

Retain only latest member rows and bounded 64-position trial ledgers. Every
arrival recomputes its whole member history, then replaces that member's old
log evidence in the global mixture. Validate scratch rows before publication.
Prediction/Issue O(HQ), arrival O(HQL), memory O(MHQ+ML); M<=200, H=3/28,
Q=21,L<=64. This is bounded, NOT constant in arbitrary corpus/history.
Narrow density s2/s4 must match frozen V40 raw shared laws through delayed
and censored evidence; this falsifies unintended law changes from new caching.

## Prospective Stages

DIAGNOSTIC seed2026104101: 28 worlds (14 regimes*2 geometries), one per cell.
NO intervals, model selection, tuning or adoption verdict. Check integration
and feasibility. DESIGN2026104103 and untouched CONFIRMATION2026104104:
16 worlds/cell,448 worlds/split. All eight candidates assessed on BOTH, no
changes after diagnostic/normal results. Verify exact seeds disjoint across
these stages and V39/V40. Three schedules immediate/fixed150/uniform299;
150 members*16 nominations=2400 distinct labels/world, noisy regimes intact.
Eight candidates: narrow/rich * density/moment * strength2/4; all shared.
Controls full64, adaptive-window bank, frozen V40 raw2_shared.
33 arms/world;550 snapshots/world; normal14784 arms/246400 snapshots/split.
Schedules/model copies are NOT independent observations. Truth arrays only
evaluator/scorer; issue before same-tick receipt, stable tie order, full drain.

## Unchanged Quality And Cost Screens

Keep original V40/V39 whole/priority issued expected Brier, final whole/
priority Brier/top10 usefulness, packet bias, two consecutive in-phase
recovery rounds(Brier<=.20,usefulness>=.75); miss=phase length+1. No global
phase recovery test for gradual/asynchronous streams. Mean +/-3.5SE on
WORLD-level paired differences, exploratory NOT simultaneous coverage.
Require all84 cells in BOTH normal splits for each candidate:
- stationary first4 plus stationary_noise10: versus Full five gain lowers>=-.01;
- nine shifted regimes: issued whole AND priority gain means>=.01/lowers>0,
  final three gain lowers>=-.01;
- versus Adaptive all regimes, five gain lowers>=-.01;
- declared discrete changed phases: recovery gain lower>0 and mean>=10%Full.

Report paired MomentMinusDensity, RichMinusNarrow, Strength2Minus4 and
MomentOnlyStrength2Minus4 contrasts within each world then across worlds.
These are not independent cell votes or omnibus significance tests.
Same learner-loop 400ms/2400-label and constructor8MiB screens. New timers
include output-array creation; old controls retain their frozen exclusion
of those initial output allocations. Report this asymmetry; do not claim
identical total-cost accounting. Both exclude scoring/serialization/audit,
acquisition/storage/serving. This does NOT establish equal TOTAL observation
cost or loaded100/250ms serving/freshness goals. Rich moment constructor,
Predict150 and fully-observed64 replay measured three times; checkpoint
restore outside replay timing. Reset peak/multi-model deployment unmeasured.

## Preflight And Audit

Freeze dependencies, old sources, new packages, tests, protocol and runner
before outcomes. Exclusive-create0600, fsync artifacts, preserve failures;
no overwrite/resampling or gate relaxation. Race/unit: mean/KL/reference/
explicit21^3 path/history/censoring/owner/epoch/cap/atomicity. Integration:
future-prefix plus independent full-history raw auditor; >=13 nonidentity
semantic corruptions and missing-allocation rejection. Vet, three benchmark
repetitions. Full arithmetic rerun is auxiliary and distinct from independent
prior/full-history law verification. Offline audit ceiling60min is declared
prospectively, separate from400ms learner cap; broader H raises audit work.
Normal results require untouched-source comparison with diagnostic freeze.

All seven whole goals stay OPEN unless their complete original criteria are
proved. This only tests model/learning components. No production/private
corpus/whitepaper/commit/push/branch/instruction changes. Newer unrelated
worktree edits are preserved, not adopted into this study or reverted.
