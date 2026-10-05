# Packet prior v25: frozen geometry and coherent forecast study

Date: 2026-10-02. This research-only Goals 1/4/5/6 component follows v24's
failed empirical protection screen. Keep its failure and exact-model
diagnostic unchanged. Freeze larger fresh cohorts, two vector geometries and
a calibrated-law control before inspecting either new split. No serving or
whitepaper change is authorized by a passed component.

## Population and evidence

Use fresh temporary actual LibraVDB/SQLite instances and the current service,
with 200 eligible and 17 future events, 256D unit vectors and a unit query.
Geometry tight uses eligible angle step .005; wide uses .020. Angles may pass
pi in the wide eligible tail: nominate by actual cosine, not event number.
Require every nominated durable norm, cosine and service baseline to match
the angle oracle as in v24. Commit exactly 150 distinct activated certified
nominees at origin+2 seconds before any outcomes. The first 32 are the fixed
monitored set. Observe one ordinary full-stream outcome per monitored event,
then Recall at origin+5 seconds with the same epoch and nominee set. All
nominees get a hidden law; hidden p never enters service or proposal code.

Run 32 independent label worlds per geometry/regime in each split: 256 worlds
and 8192 accepted outcomes per split. Design base seed is 2026102503 and
confirmation is 2026102504. Add geometry_index*10000000, regime_index*1000000
and world*1000. Hidden independent-law permutations use a separate RNG stream
with offset 5000000000. Regime order is independent, aligned, reversed,
calibrated. Before labels, set p in prefeedback journal order:

- independent: balanced random permutation of 75 .8 and 75 .2 probabilities;
- aligned: p_j=.9-.8*j/149;
- reversed: p_j=.1+.8*j/149;
- calibrated: p_j equals that event's committed prefeedback BaseLaw.Useful.

The last case gives the initial probabilistic baseline its declared true law;
it is a synthetic control, not a claim that production retrieval scores are
calibrated. Preserve the prefeedback prior mean b_i^ref separately from the
later same-as-of baseline b_i. Synthetic certificates assume coverage only.
Each monitored event contributes all its single observed outcomes with unit
weight. The monitored set is rank-dependent; no population selection guarantee
or external causal evidence follows.

## Models and rank comparisons

Research module `researchbeta` declares one Beta-Bernoulli joint family:
theta_i~Beta(s*m_i,s*(1-m_i)); training Y_i and a future Y_i' are independent
Bernoulli(theta_i) conditional on theta_i. Both the evidence likelihood and
outcome kernel come from this same family. Freeze s=2. Flat uses m_i=.5;
context uses m_i=b_i^ref. With accepted ordinary counts u_i,v_i, posterior
predictive is (s*m_i+u_i)/(s+u_i+v_i). The model is static and per-event;
it supplies no changepoint, source-independence or epoch certificate.

Conjugate Beta updates are established mathematics, not a new inference
algorithm; see Diaconis & Ylvisaker (1979), [Conjugate Priors for Exponential
Families](https://www.cs.columbia.edu/~blei/fogm/2020F/readings/DiaconisYlvisaker1979.pdf).
The choice of contextual anchor, frozen strength and EventFrame rank use are
the research proposals tested here; the source confers no performance or
external-law guarantee on them.

Compare the actual journaled law with two offline proposed laws. Monitored
members use the full flat or contextual posterior predictive; unmonitored
members retain b_i. These proposed probabilities are recorded separately
and are never written into the actual service forecast bundle. Score all
150 probabilities against their hidden future law with expected Brier.

Separately replay four rank controls at recall_k=50, pack_k=10, token budget
10000 and the current packing policy:

1. baseline: b_i;
2. current: actual committed RankScore;
3. flat innovation: clip(b_i+.1*(flat_predictive_i-.5));
4. context innovation: clip(b_i+.1*(context_predictive_i-b_i^ref)).

Unmonitored candidates keep b_i; clips are to [0,1]. The rank controls keep
actual scored bundles unchanged so packet-quality gains cannot be mistaken
for forecast-law improvement. Require current replay to match actual packet
IDs exactly. Report whether the two proposed packets differ, all-candidate
and packed expected Brier, expected usefulness and false-item counts. Record
accepted sufficient counts, priors, hidden laws, all probabilities and scores.

## Frozen screens

Use paired per-world differences with mean +/-3.5 SE across the 32 independent
worlds in each geometry/regime. These are descriptive screening intervals,
not simultaneous confidence sequences. Require all integrity and no-future
checks; preserve every failed gate and keep the two screens separate.

Rank proposal screen, separately in both splits and geometries:

- independent and reversed each need mean packet usefulness gain >=.02 over
  baseline and a positive lower endpoint;
- aligned and calibrated each need harm upper endpoint <=.01.

Forecast proposal screen, separately in both splits and geometries:

- independent and reversed each need mean whole-frontier expected-Brier gain
  >=.005 over baseline and a positive lower endpoint;
- aligned and calibrated each need Brier-harm upper endpoint <=.01.

Do not tune s, rank weight, monitored set, sample size or thresholds on these
cohorts. A rank or forecast screen pass proves only this finite synthetic
component. Shifts, correlated/delayed evidence, prospective agent outcomes,
integration with residual calibration, visible writes and loaded freshness
remain requirements of the seven whole goals.

Run independent source/data-hash, probability, packet and risk reconstruction.
Run ordinary and race controls for the pure module and six full service
correctness worlds (both geometries and independent/aligned/calibrated) with
separate diagnostic seeds. Those six add no independent statistical worlds.
Record sequential service times and module arithmetic throughput/memory;
exclude any claim of loaded p99 from those timings.
