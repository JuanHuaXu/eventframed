# Repeated-trial dispersion V34: frozen protocol

Frozen2026-10-03 before outcome collection. All seven whole goals OPEN.
No production, earlier evidence, priors or whitepaper changes.

## Population and Contract

Seed bases2026103403 design and2026103404 confirmation. Two cosine
geometries,12 declared regimes,32 worlds/cell yield768 worlds/split.
World seed=base+geometry*100million+regime*1million+world*1000.
Truth/label/nomination RNG offsets101/202/303. Cosine baselines match V30;
new implementation and explicit generator checks replace its legacy runs.

Regimes, in frozen order: aligned,reversed,calibrated,curved,independent,
permuted_curved,narrow_peak,baseline_matched,mean_half,mean8,mean32,
mean_shared. First seven are deterministic rates as V30 defined them;
independent permutes75 rates .8 and75 rates .2. Baseline_matched draws
phi~Beta(2b,2(1-b)). Last four use mean p=(.9b-.05)/.8 with persistent
phi~Beta(kappa*p,kappa*(1-p)) at kappa=.5,8,32,infinity respectively.
Infinity means phi=p. Truth knows phi; learners only see b and arrived
binary outcomes. This does not broaden the mean family after evaluation.

Each world draws16 rounds of150 genuine conditionally independent trials,
using one persistent phi/member. A label-independent RNG permutation per
round nominates every member once; conditional nomination probability
1/(150-position). All four learners receive the same sequence:
adaptive135-state mixture, fixed2, shared-only and independent local Beta2.
Static baseline is a fifth forecast. Trial ordinal rejects duplicates;
these are independently generated trials, NOT transcript replay or copies.
Issue all four forecasts before accessing the corresponding outcome.

Checkpoints32,150,300,600,2400 labels are fixed. Score expected future
Bernoulli risk at the150 existing members, whole/priority Brier, top10
usefulness, packet Brier/bias. Priority weights3 for first10,1 elsewhere,
denominator170. No observed label is scored as a future realized outcome.
Record exact trial tape, issued laws, snapshots and dispersion weights.

## Frozen Component Gate

Primary screen at2400 labels, both splits:

- Every cell: adaptive protects local and fixed2 whole/priority Brier and
  usefulness with paired mean-3.5SE>=-.01.
- Low-variance exact mean cases aligned,reversed,calibrated,mean_shared:
  whole/priority gain versus fixed2 >=.005 and mean-3.5SE>0.
- Every cell: packed bias abs(mean)+3.5SE<=.10.
- Zero mathematical/replay/future/source errors; per-model measured phase
  sums <=25ms. These sums are NOT wall/serving latency.

Other budgets diagnose sample efficiency, not a data-selected adoption gate.
Shared-only is retained as a diagnostic, not a universal safety baseline.
Intervals over32 independent worlds/cell are not confidence sequences or
simultaneous AP certificates. Checkpoints within a world are dependent.
A passed2400-label screen would be a fixed-world dispersion component,
NOT completion of a sample-efficient shifted challenger or any other goal.

## Audits and Cost

Freeze model/tests/generator/protocol hashes in raw manifests. Exclusive
output creation prevents accidental replacement. At checkpoints a separate
batch beta-integral computation verifies full joint weights and laws; exact
RNG replay checks every generated rate, outcome, nominee and issued law.
Validate generator formulas and all denominators/packets independently.
Corrupted-law/weight/outcome/ordinal/source controls must reject altered
copies. Flip unarrived outcomes and prove earlier records are unaffected.

Record per-model setup, cumulative pre-outcome probe, update, current and
earlier snapshot times. Nomination/elapsed are separate shared/interleaved
work. No acquisition, database, queue or loaded agent cost is hidden inside
these components. No equal-total-cost Goal7 or delayed/temporal claim.
