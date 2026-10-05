# Prior And Borrowing V40 Prospective Protocol

2026-10-03. Isolated research, frozen BEFORE diagnostic/normal outcomes.
V39 failed all3candidate models in78/84cells on BOTH splits. Geometry-based
overconfidence, uncalibrated centers, lost borrowing and reset misspecification
are competing explanations; none proven as the sole cause. New factorial
tests three first explanations without editing prior sources or production.
No upstream daemon bug or production patch is asserted.

## Research And Joint Model

[Adams and MacKay2007](https://www.cs.princeton.edu/~rpa/pubs/adams2007changepoint.pdf)
separates changepoint dynamics from predictive models and integrates parameter
uncertainty. [Knoblauch and Damoulas2018](https://proceedings.mlr.press/v80/knoblauch18a/knoblauch18a.pdf)
integrates online model uncertainty with changepoint prediction. Those papers
motivate declared priors and model averaging, NOT prove our borrowing scheme
or give its error guarantees. Our finite calibration hierarchy is a NEW
research approximation, NOT their run-length/VAR algorithms or exact
continuous Beta BOCPD. No published speedup/detection theorem is inherited.
Wilson/Nassar/Gold2010 hazard-learning was located; full PMC reader was blocked
and is not used to derive this implementation. Retain unknown-hazard learning
as a separate possible lead. A guessed proceedings PDF was inaccessible;
no mathematical claim relies on it.

For baseline b_i, declare c_i=b_i(raw) or(.9b_i-.05)/.8(inverse). The latter is
a previously declared research family, NOT empirically calibrated truth or
an oracle-valued prior. Let mu_i,h=(c_i,1-c_i,1/2), h=0,1,2, and omega=(.8,.1,.1).
Rate atoms a_z=(z+1/2)/21. Conditional prior:

```math
pi_i,h(z) proportional to a_z^(s*mu_i,h-1) * (1-a_z)^(s*(1-mu_i,h)-1)
T_i,h(z'|z) = (1-lambda)*1[z'=z] + lambda*pi_i,h(z')
```

s=2or4;lambda=1/16 per member nomination, unchanged across candidates.
Normalize discrete masses EXACTLY. This is a Beta-SHAPED finite prior;
its mean need not equal mu, and s is NOT an exact continuous pseudocount.
Shared mode has ONE calibration family h drawn with prior omega; private
mode has independent h_i with SAME prior. Member rate chains are conditionally
independent; shared family does NOT force equal means/rates or grant AP merges.

For as-of arrived emissions E_i and nomination count n_i, integrate full
member chain paths to marginal likelihood L_i,h. Unknown/pending/cancelled
emissions are1; observed labels have Bernoulli(a_z) likelihood. Prediction
and evidence likelihood are marginals of this same joint finite model:

```math
W_h shared proportional to omega_h * product_i L_i,h
W_i,h private proportional to omega_h * L_i,h
q_i,next = sum_h W_h-or-W_i,h * E[a_next | h,E_i,n_i]
```

Shared/private start with IDENTICAL marginals; evidence from another member
changes only the shared alternative. Local latent rates still adapt separately.
The new three-family prior also changes initial marginal versus V39's single
family; raw4_private is NOT an exact V39 replay. Legacy local16 is retained
to expose that difference. Only matched private/shared contrasts isolate
cross-member borrowing within this declared hierarchy.

Late evidence replays its original member suffix, replaces OLD marginal
likelihood contribution with NEW, never multiplies old labels twice. Global
weights use log evidence; all proposed rows/weights validate privately before
publication. Ticket owner/epoch/ordinal/time and original privately retained
forecast survive delay/replay/cancellation. Epoch reset is explicit, not a
changepoint decision. Single-owner, not shared-goroutine safe. Missingness
assumes non-informative selection/censoring; independent fixed/uniform delays
satisfy this experiment, NOT arbitrary outcome-dependent production evidence.

## Prospective Stages And Complete Scope

StageA DIAGNOSTIC seed2026104001:2geometries*14regimes*1world=28worlds,
three delay schedules,11models=924arms. This is NOT confirmation/FIT data,
has NO within-cell interval/adoption verdict, and cannot tune any candidate.
Purpose: complete-path integration, feasibility and descriptive ablations.

StageB DESIGN2026104003 and CONFIRMATION2026104004,16worlds/cell per split:
448worlds/split,14784arms/split. All8candidates evaluated on BOTH splits,
no best-model selection/tuning or new priors after outputs. Source IDs/RNG
checked against other stages; cohort seeds distinct from V39. Each world
has550snapshots(11*(16+17+17)), so246400/split,492800total.
Each world has2400distinct labels(150members*16nominations); model copies
and schedules are NOT additional independent observations. Diagnostic adds
67200distinct labels and15400snapshots. Intervals computed on independent
WORLD-level paired differences, not arms/rows as if independent.

Eight candidates:raw/inverse *strength2/4 *private/shared. Three controls:
full64, adaptive-window bank, legacy local16. ALL get identical nominations
and evidence labels. Keep ALL14regimes, two geometries and schedules
immediate/fixed150/uniform299.10% feedback noise in two regimes stays hidden
from learners; true rate arrays only evaluator/scoring. Issue before same-tick
arrival, nomination order for ties, drain all labels. No future truth to model.

## Unchanged Normal Screens

Retain V39 original issued expected Brier and priority(first10weight3), final
whole/priority Brier/top10utility, packet bias, and two consecutive in-phase
recovery rounds(Brier<=.20,usefulness>=.75); miss=phase length+1. No global
phase test for gradual/asynchronous. Every normal cell has16paired worlds;
report mean +/-3.5SE exploratory intervals, NOT simultaneous confidence
sequences, Anti-Pigeon coverage, or calibrated error guarantees.

For EACH8candidate, require ALL84cells in BOTH normal splits:
- versus FULL: first4stationary plus stationary_noise10 all5metric lower>=-.01;
  shifted9 issued whole AND priority mean gain>=.01/lower>0; final3 lower>=-.01;
- versus ADAPTIVE: every regime, all5metric lower>=-.01;
- discrete changed phases: recovery lower>0 and mean>=10%FULL control mean.

Paired descriptive main effects for center, strength and borrowing are averaged
across4matched combinations WITHIN each world, then interval across16worlds.
No diagnostic pseudo-CIs, no multiple-model confirmation selection. Record
all per-cell metrics/gates; endpoint improvement cannot rescue issued-law harm.

Same complete collector work cap400ms/2400labels and constructor8MiB asV39.
Includes schedule/setup/issue/arrival/drain/snapshots/receipt-loop overhead,
excludes scoring/serialization/auditing/acquisition/durability/serving. It is
NOT equal TOTAL acquisition/observation cost or loaded100/250ms daemon proof.
Candidate costs O(HQ)Predict/Issue, O(HQ*L)late replay, memory O(MHQ*L), H3,Q21,
M<=200,L<=64. Bound is scoped, NOT constant in unbounded corpus/history.
Measure constructor150, Predict150 and late64FULLY OBSERVED suffix(63known
later labels, not old unit-emission benchmark),3repetitions. Checkpoint restore
outside replay timer; report excluded setup honestly. Epoch-reset peak memory
and simultaneous multiple-model deployment are NOT measured by constructor.

## Preflight And Audit

Freeze all participating old/new sources, prospective runner and independent
whole-history reference BEFORE StageA. Exclusive-create0600 raw/manifests/logs,
fsync, preserve all failures; no overwrite/resampling or silent gate changes.
Race:finite model/path/history/delayed/interleaved/private/shared/owner/epoch/
atomicity/cap; collector future-prefix. Vet and3benchmarks. Technical auditor
unrelated seed730007 with non-identity semantic corruptions/missing-allocation
rejection before outputs. Independent reference uses unnormalized FULL history
integration, not normalized cached suffix. Whole-source/RNG/population/domain,
original forecasts/receipts/snapshots/metrics/recovery/coverage/timing verified.
Old FULL/ADAPTIVE/local controls retain their independently audited definitions.
Auditor may also exact-rerun auxiliary arrays, clearly distinct from independent
math validation. Slow offline audit has35minute ceiling separate from400ms
learner cap; filename metadata ends-command.json, NEVER collides with -audit.json.

All seven whole goals OPEN. This only tests model/learning components; no
Anti-Pigeon authority, source authenticity, real-agent answer, causal inference,
falsification acquisition superiority, or continuous-serving guarantee.
No production/private corpus/whitepaper/branches/commit/push/instruction changes.
