# Learner refinement v5 protocol

Frozen 2026-09-12 before any v5 design/confirmation run. First work package for
research-direction.md recommendations 1,2,4. No tuning between splits.

## Boundary

This is a learner-level, fully observed synthetic EventFrame experiment. Each
prediction reads all nine coordinates via the existing EventReader. It does NOT
test the six-coordinate MMM observation allocator, reference sharing or AP gates.
Do not compare absolute scores to v3 as if observation budgets were unchanged.
The incumbent and bayes.ForecastMix are reused; no production defaults change.

Seven arms: frozen incumbent; short-only; adaptive-only; fixed mixture
[incumbent,short64,long256,uniform]; adaptive mixture
[incumbent,adaptive,long256,uniform]; forest mixture
[incumbent,forest,long256,uniform]; combined mixture
[incumbent,adaptive,forest,uniform]. Same .7/.1/.1/.1 prior and .002 share as v3.
No incumbent or mixture resets. Standalone arms expose the contribution of
preservation/mixing rather than assuming every learner gain needs a mixture.

Audit selection is independent Bernoulli(.25), fixed before outcome. Fixed models
fit at32 admitted audits, then every16 on latest64/256. Every arm has the same
audit evidence. Initial incumbent fits4096 examples from its own fitting seed.

## Adaptive prototype

Monitor journaled incumbent Brier losses on available full-stream outcomes.
Keep at most256 loss/time pairs. Every8 available outcomes, scan cuts with >=16
samples per side. For each cut n0+n1=n, compare absolute mean gap with
sqrt(.5*(1/n0+1/n1)*log(4*n/.01)). If any pass, keep the suffix at the greatest
gap-minus-bound cut; at most one cut per update. Restrict adaptive training to
audits at/after the cutoff, capped256. Refit on cut or each16 audits, minimum32;
otherwise it forecasts .5 until supported. The incumbent stays intact.
This is a bounded ADWIN-INSPIRED mean-change heuristic, not faithful ADWIN2,
not its logarithmic implementation, and not an anytime certificate.

## Tree prototype

Five incremental binary trees, depth cap4, six randomly selected candidate fields
per leaf, no repeated path field. Independently Poisson(1)-resample each admitted
audit, truncated at8 copies; update each copy sequentially. Leaves maintain
class/field counts. At >=32 copies and every16, split if positive best information
gain exceeds second-best by sqrt(log(1/.05)/(2*n)), or that bound is <.1.
Leaf forecasts have Beta(1,1) smoothing; ensemble averages five probabilities.
New children start empty; no automatic reset, replacement, or retrospective
redistribution of training examples. One forest per trajectory, independent RNG.
This bounded streaming ensemble is NOT a full Adaptive Random Forest and imports
no split-confidence guarantee for repeated checks or resampled observations.
Interaction-only shifts are a deliberate hard negative for greedy univariate splits.

## Evaluation matrix

Ten cases: stable noise.05; stable noise.20; abrupt shifts at128,256,384 noise.05;
gradual transition128..384 noise.05; recurring every128 noise.05; shift256 with
fixed16-step feedback delay plus independent25% missing labels; local three-bit
XOR shift256; null fair labels. Other old law is process three-bit XOR; ordinary
new law is local bit2. Scenarios never enter the learner or detector.

Three independent fitting seeds per case, eight evaluation trajectories per fit,
two splits: 480 streams of512 predictions, seven paired arms. Fit seeds derive
from 2026092301*1000000 + scenario*1000 + fit_index. Evaluation seeds derive from
split_base*1000000 + scenario*10000 + fit_index*1000 + trajectory*10 + role;
split bases2026092302/2026092303, roles0 generation,1 audit,2 missing,3 forest.
All are disjoint from previous pilots. Generate all labels for external evaluation,
but missing labels NEVER enter learners/monitor/weights. Delayed labels arrive
only after the due step's forecast. Record originating forecast and availability.
After step511, do not flush pending feedback into evaluated predictions.

## Criteria

For each new mixture versus fixed mixture, report full/post Brier and paired
intervals in every case, plus accuracy, log loss, recovery windows and costs.
Positive refinement: gain >=.005 and lower bound>0 on both shift128 and shift256,
stationary harm upper bound<=.01 in both noise conditions, and no >.01 mean harm
in any case/window. Check each fitting-seed mean separately for sign consistency
on the two primary shifts; do not claim robustness if one reverses.

Use z=3.6 paired normal intervals across24 evaluation trajectories conditional on
the fixed set of3 fitting models; 3 mixtures x10cases x2windows x2splits=120
contrasts. These are approximate, not unconditional training-seed coverage.
Post starts at change (128 for gradual/recurring,256 for stable/null diagnostics).
Also preserve standalone-arm scores; per-fit results; monitoring cuts; supplied
label counts; audit counts; unresolved queued outcomes; forest node counts;
and 80%-correct rolling64 recovery completion/misses on abrupt shifts only.
No pooled average can conceal a failed protection case.

Run unit/race checks before evaluation; replay and verify artifacts afterward.
Benchmark model prediction/update/refit separately, serially after experiments.
Sources: ADWIN and ARF links in research-direction.md; this implementation states
its departures explicitly. A passing result licenses another research stage,
not partial-view MMM integration or deployment.
