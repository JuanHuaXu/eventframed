# Calibration observation v27: frozen challenger and acquisition study

2026-10-02. V26 improved retrieval but left large selected-packet probability
bias. Preserve its data. Test a distinct bounded research model offline before
any served-law integration. No production or whitepaper changes.

## One Joint Model

Fixed frontier n=150, maximum200. Sorted baseline b is the cosine oracle
`.65*(cos(angle)+1)/2+.275`:198 eligible angles step*(i+1), plus two zero-angle
events; take the best150. Tight step=.005, wide=.020. This approximates the
earlier fixture's geometry, not an actual new Service execution.

There are27 hypotheses. Hypothesis0 has p=b, weight.1; hypothesis1 has
p=(.9*b-.05)/.8, weight.8. The other25 have
p=clip(a+c*r,.02,.98), weight.004 each, with
a in{.1,.3,.5,.7,.9}, c in{-.8,-.4,0,.4,.8}, r=i/(n-1).
The affine pairs have mean.5 by complement symmetry, so the prior predictive
is exactly b. Reject b outside[.25,.925] and unsupported frontier sizes.

Draw theta from this prior; conditional on theta, distinct event rates are
independent phi_i~Beta(2*p_theta(i),2*(1-p_theta(i))). Past and future labels
for one event are independent Bernoulli(phi_i) given phi_i. The evidence
likelihood and outcome predictive are marginals of this same joint family.
Posterior theta weights update by the arrived event's predictive likelihood;
event sufficient counts update separately. There is one label per event here.
Predictive q_i=sum_theta w_theta*(2*p_theta(i)+u_i)/(2+u_i+v_i).

This is ordinary Bayes under a declared finite model, not a truth certificate
or Anti-Pigeon sharing authorization. The complementary affine family is
motivated by the existing benchmark; explicitly retain off-family controls.
Do not claim external generalization from fitting these known geometries.

## Observation Policies and Cost

Compare same-model head, random without replacement, uncertainty maxH(q),
information max[H(q)-sum wH(p_theta)], and fixed bit-reversal stratification.
Selection receives no hidden truth or undelivered label. For unobserved nodes,
the information formula is exact I(theta;Y_i|history), not globally optimal
acquisition. Adaptive design factors cancel conditional on past history when
labels satisfy this joint model; this does not fix retrieval-selection bias
or source dependence in the actual daemon.

Also retain V26's local per-event Beta model with head32 as the fixed-label
control. All policies use the same potential-label tape within a world.
Unobserved individual rates and outcomes never enter model/acquisition code.

Run fixed32 labels and modeled total-cost caps for label costsL=10000,100000,
1000000 units. Setup cost=20*n*M; reserve final=3*n*M+n*ceil(log2(n)).
Each update costs5*M. Head/random scan cost=n+1; stratified=nextpow2(n);
uncertainty cost=n*(3*M+12)+n; information=n*(6*M+12)+n.
Budget=setup+final+32*(L+5*M+information_cost). Use at most n observations;
do not take an action when its selection+label+update and reserved final work
would exceed the cap. Other policies may buy more labels from saved compute.
Report actual counts and unused budget. These are explicit modeled cost units,
not measured FLOPs or an empirical equal-wall-time/agent-acquisition guarantee.
Measure preparation, selection, update, final prediction and total CPU times
separately; the actual-cost Goal7 remains open without external acquisition.

## Cohorts and Frozen Screens

Use32 worlds per geometry/regime and fresh bases2026102703 design and
2026102704 confirmation, offsets geometry*10000000+regime*1000000+world*1000.
Independent balanced75 rates.8/.2; aligned=.9-.8*r; reversed=.1+.8*r;
calibrated=b; smooth=.1+.8*sin(pi*r)^2; matched draws theta and then phi
from the declared hierarchy. Separate truth and label RNG streams. Record
all rates, potential labels, chosen indices, pre-label forecasts, final laws,
weights, metrics, modeled costs, timings and source hashes.

Fixed-label rescue screens for stratified and information, both splits and
geometries: independent/reversed whole-frontier and priority-weighted Brier
gain>=.02 versus local control, positive paired lower endpoint; aligned and
calibrated harm upper<=.01 for these scores and packet usefulness.
Priority weights3 for baseline top10,1 otherwise, normalized. Mean packet
probability bias magnitude upper=abs(mean bias)+3.5SE must be<=.10 in every
regime, including smooth and matched. This is not complete conditional
calibration. Preserve all failed gates.

Modeled-cost observation screen at L=100000: information must improve Brier
by>=.005 and packet usefulness by>=.02 over both random and uncertainty on
reversed/smooth/matched, with positive paired lower endpoints; protection
harm upper<=.01 on independent/aligned/calibrated. Report other L as cost
sensitivity. Intervals mean+/-3.5SE over32 worlds are descriptive, not
simultaneous certificates or sequential stopping authority.

All seven whole goals remain open unless their broader requirements are
achieved. This study has no shifts, source-dependence authority or actual
agent responses. Sources: [Houlsby et al. (2011)](https://arxiv.org/abs/1112.5745)
for predictive-entropy information acquisition;
[Cockayne et al. (2022)](https://jmlr.org/papers/v23/21-1065.html) motivates explicit calibration
testing, not a theorem for this experiment.
[Safe-Bayesian GLMs (2020)](https://proceedings.mlr.press/v108/heide20a.html) documents misspecification
as a distinct concern; no SafeBayes algorithm or guarantee is implemented.
