# Nonlinear partition v29: frozen empirical screen

2026-10-02. Freeze before collecting design or confirmation. Retain the
preflight model: .99 independent baseline branch, .01 total prior over26
depth3 dyadic partitions, uniform Beta leaf means, strength2 individual rates.
Do not change these priors or depth after seeing outcomes. Model groups are
hypotheses, not Anti-Pigeon authority. No production, paper or dependency edits.

## Matched Input and Controls

150 events from each v27 cosine geometry: two zero angles plus198 positive
angles with tight step .005 or wide .020, sorted descending by baseline
`.65*(cos(angle)+1)/2+.275`, retaining150. Coordinate=i/149, fixed before
labels. This is a normalized frontier-position proxy, not a semantic or causal
representation. All arms share a world's potential-label tape.

Fixed32 arms: original unchanged baseline; local Beta head32; old v27
stratified/information; new partition head/random/stratified/uncertainty.
The primary challenger is partition-stratified. Compare it with BOTH local
and old-stratified. Old/new stratified select the same indices in the same
order: this isolates the model change at equal evidence, rather than choosing
a favorable acquisition policy afterward. One label per event, prediction
before revelation. No truth rates, oracle labels or unselected outcomes in
the model/selector input. Freeze every source hash in the manifests.

## Fresh Worlds

32 worlds per geometry/regime,576 worlds per split. Fresh bases2026102903
design and2026102904 confirmation, offsets geometry*10000000+
regime*1000000+world*1000. Independent truth/label/acquisition RNG streams
use+101/+202/+303. Regimes, in fixed order:

- independent: randomly assign75 rates .8 and75 rates .2;
- aligned: .9-.8*r;
- reversed: .1+.8*r;
- calibrated: b;
- curved: .1+.8*sin(pi*r)^2, consumed shape but fresh outcomes;
- shifted_peak: .1+.8*exp(-((r-center)/.14)^2),
  center=.25+.5*(world%8)/7;
- alternating: .8 at even member positions,.2 at odd positions, finer than
  the partition resolution;
- baseline_matched: independent phi~Beta(2*b,2*(1-b));
- tree_matched: sample a partition from its normalized conditional tree
  prior, independent psi_leaf~Uniform(0,1), then independent
  phi_i~Beta(2*psi_leaf,2*(1-psi_leaf)). This is conditioned on the .01 tree
  branch, NOT an unconditional draw from the full .99/.01 model.

The last two regimes separately test the two branches. Tree leaf means and
member rates remain hidden from inference. Record all sampled truth values,
training labels, indices, pre-label forecasts, final laws/weights and packets.
Report Beta/Gamma floating-point sampling limits rather than claiming exact
real-number random draws at extreme parameters.

## Unchanged Strength of the Gates

Primary fixed32 screen, BOTH splits and geometries: in independent/reversed,
whole-frontier and priority-weighted Brier gain>=.02 over BOTH controls, with
positive paired lower endpoints. In curved/shifted_peak, these gains>=.01
over old-stratified plus packet usefulness gain>=.02, with positive lower
endpoints; relative to local require harm upper<=.01 for risk and packet.
On aligned/calibrated/alternating/baseline_matched/tree_matched, harm upper<=.01
for risk, priority risk and packet usefulness versus both controls. Every
regime must have packed-bias magnitude upper=abs(mean)+3.5SE<=.10.
Priorities3 for original baseline top10,1 otherwise, normalized over170.
Intervals over32 paired worlds are descriptive, not simultaneous certificates.

## Cost and Runtime

Also compare each model/policy at shared modeled total caps for label
costs100000 and1000000. Setup=20*n*27, final reserve=3*n*27+n*ceil(log2(n)),
update=5*27+16 for both model families. Head/random selection=n+1;
stratified=nextpow2(n)*ceil(log2(nextpow2(n))) to explicitly charge bit work;
uncertainty=n*(3*27+12)+n; old information=n*(6*27+12)+n.
Cap=setup+final+32*(labelCost+update+oldInformationCost); at most150 labels.
Reserve final work before admitting another observation. These are stated
accounting assumptions, not a hardware or conservative-operation theorem.
Source-of-label costs remain hypothetical. Record actual label counts,
construction/selection/update/final prediction elapsed time and model-call
total. Measure allocation/latency separately; exclude data generation,
truth scoring, packet sorting, database, queues and serving from model timings.

The cost tables are sensitivity evidence for Goals1/4, NOT a Goal7 success
screen: fixed stratification is not falsification-driven observation.
Even a component pass cannot complete agent outcomes, AP error control,
temporal recovery or loaded freshness. All seven whole goals remain OPEN.
