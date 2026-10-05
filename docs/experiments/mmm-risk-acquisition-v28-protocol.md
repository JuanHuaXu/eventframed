# Target-risk acquisition v28: frozen fresh-cohort protocol

2026-10-02. Isolate acquisition, not model tuning: retain v27's27-hypothesis
hierarchical Bernoulli family, prior predictive identity,150-event frontiers,
two cosine geometries and six truth regimes. No changes to the v27 artifacts,
production, whitepaper or scored Service. V28's Gram implementation and exact
two-outcome identity are described in the preflight. Freeze before either run.

## Intervention and Controls

Same-model policies: random, uncertainty, information, and risk. Risk selects
the maximal expected reduction in weighted future Brier risk, with target
priority3 for original baseline top10 and1 otherwise, normalized. Include
the within-member Beta variance, not only theta covariance. These priorities
are predeclared and do not consult truth or outcomes. Keep v27's local per-event
Beta head32 fixed-label control. One distinct label per event; no repeat queries.
All arms in one world use the same potential-label tape, with forecasts issued
before selected labels are revealed. The model/selector sees no truth rates.

## Acquisition Cost

Fixed32-label comparisons plus identical total modeled caps at label costs
10000,100000,1000000. Primary cost screen remains100000, not the most favorable
cost after inspection. Setup=20*n*M; final reserve=3*n*M+n*ceil(log2(n));
update=5*M. Existing policy selection charges match v27. Risk charges
6*n*M*M+12*n*M+40*n=710700 units at n150,M27, explicitly paying for its Gram
matrix and quadratic forms. These are accounting assumptions, not measured
CPU instructions, hardware time or a proved conservative operation bound.
The component benchmark found risk approximately16 times slower than information.

Budget=setup+final+32*(labelCost+update+riskSelectionCost). Every policy uses
this same cap, reserves final work before admission, and stops before any
selection+update+label would overrun it. At most150 labels. Saved compute may
buy cheaper policies more labels; do not hide that advantage. Record used
and unused cost, labels, construction/selection/update/final prediction elapsed
times and total model-call elapsed time. Timings exclude generation, oracle
scoring, packet sorting, database, label acquisition, queues and actual serving.

## Untouched Cohorts and Gates

32 worlds per geometry/regime:384 worlds per split. Fresh seed bases
2026102803 design and2026102804 confirmation; geometry*10000000+
regime*1000000+world*1000 offsets. Truth stream+101, label stream+202 and
random acquisition+303. Prior/theta and Beta sampling match the v27 generator;
truth regimes are independent balanced .8/.2, aligned, reversed, calibrated,
curved off-model and hierarchy-matched. Record theta/rates/labels separately
from model input, all selection traces, weights, laws, packets and source hashes.

Fixed32 rescue: compare risk against the local control. In independent and
reversed cases, whole-frontier and priority-weighted Brier gain>=.02 with
positive paired lower endpoints. On aligned/calibrated cases, harm upper<=.01
for both risks and packet usefulness. All six regimes must have packet-bias
magnitude upper=abs(mean bias)+3.5SE<=.10. Both splits and geometries must pass.

Primary modeled-cost Goal7 screen: at label cost100000 risk must beat BOTH
random and uncertainty on reversed/curved/matched regimes by whole-frontier
Brier gain>=.005 and packet usefulness gain>=.02, with positive paired lower
endpoints. Independent/aligned/calibrated risk and packet harm upper<=.01.
Information is an additional mechanistic comparator; preserve its outcomes.
Other costs are sensitivity, not replacement screens. Report mean+/-3.5SE
over32 paired worlds; these are not simultaneous confidence sequences.

Verify posterior/trace/packet/risk/cost reconstruction independently. Check
Gram selection against direct covariance sums, distinct evidence, invalid
inputs, source hashes and a future-tape falsifier. Replaying random seeds is
a separate Go check; the JavaScript verifier need not duplicate Go's RNG.

This study has no temporal shifts, source-dependence guarantee, real-agent
responses or production latency. No whole goal is completed by this finite
screen. All seven success criteria remain intact.
