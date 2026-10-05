# Packet innovation v23: frozen all-candidate screen

Date: 2026-10-02. This isolated study follows v22's missing packet-utility
denominator. It evaluates actual service forecasts and a replay of two proposed
rank corrections. No proposal receives hidden usefulness probabilities. This is
a finite synthetic screen toward Goals 5/6, not an agent-task or loaded serving
validation. The v22 runner and its consumed cohorts remain unchanged.

## Population and evidence

Use the existing 256D cosine fixture with 200 eligible and 17 future events,
fresh temporary LibraVDB/SQLite instances and current service defaults. Freeze
eight independent label worlds per regime in each of design and confirmation.
Design seed base is 2026102303; confirmation is 2026102304. Offset each regime
by 1000000 and each world by 1000. Separate RNG streams generate hidden laws and
training labels. The geometry is fixed, so independent worlds establish only
the declared label/generator comparison.

Commit an initial Recall at origin+2 seconds, requiring exactly 150 distinct,
activated, certified nominees and a durable journal. Define hidden usefulness
probabilities for **all 150 nominees**, in the journal's prefeedback order:

- independent: a fresh balanced random permutation of 75 probabilities .8 and
  75 probabilities .2;
- aligned: p_j=.9-.8*j/149 for j=0..149;
- reversed: p_j=.1+.8*j/149.

The first 32 nominees are a predeclared monitored set. Observe one independent
Bernoulli outcome for each at origin+3 seconds plus its millisecond offset.
Every outcome of each monitored event is observed; submit those event-local
full-stream observations with inclusion probability 1. Selection of the
monitored set remains rank-dependent and does not establish a population audit
or external certificate. Evaluation needs no inverse weighting because the
hidden law covers every nominee. Require distinct event posterior keys.

Recall at origin+5 seconds after feedback. Require the same 150 nominees and
evidence epoch, 32 accepted updated belief laws and no future event. Selection
and omitted-influence certificates are synthetic assumed inputs. Read actual
accepted Beta statistics solely for the monitored events; require their alpha
and beta to match one ordinary outcome added to Beta(1,1). No retagging, stale
reuse, group pooling or parameter fitting is allowed.

## Equal-evidence comparisons

Reconstruct candidates from the learned journal and stored events. First
replay current RankScore ordering, recall_k=50 and the actual packing policy;
its packet IDs must exactly match the service response. Then compare these
same-as-of controls over the same nominee population:

1. baseline: b_i, the actual BaseLaw.Useful (identity calibration in this
   fixture), ignoring learned rank deltas;
2. current: the actual journaled RankScore;
3. flat innovation: b_i+.1*(m_i-.5), where m_i is the accepted Beta mean;
4. context innovation: b_i+.1*(m_i^ctx-b_i), with
   m_i^ctx=(2*b_i+alpha_i-1)/(alpha_i+beta_i).

For both proposed controls, unmonitored events keep b_i and final scores clip
to [0,1]. The context formula replaces the two flat pseudocounts with two
pseudocounts centred on the retrieval baseline while keeping actual observed
success/failure counts. It is a proposed working rank model whose prior may be
miscalibrated, not an established scored-law replacement. Both proposals keep
the actual scored forecast bundle unchanged. Their only inputs are b_i and
accepted past Beta statistics, never p_i or future labels. Stable descending
score order uses initial journal order to break ties. Truncate to 50 before
packing 10 with token budget 10000 and the current default packing policy.

Measure mean hidden usefulness of each packed packet, expected false items
sum(1-p_i), actual journal-law expected Brier over **all 150 nominees**, and
packet expected Brier. Hidden p enters offline evaluation only. Report packet
identities and every nominee's law, scores, prior/evidence counts and p. Measure
sequential Recall/outcome durations and isolated proposal scoring time; none
of these is a loaded p99 or observation-freshness claim.

## Frozen rescue screen

Use paired per-world packet utility differences versus baseline, with mean
plus/minus 3.5 standard errors across eight worlds within each regime. These
are descriptive finite-screen intervals, not simultaneous confidence sequences
or guarantees over the external target law. A proposal passes this preliminary
screen only if, separately in both splits:

- independent and reversed each have mean utility gain at least .02 and a
  strictly positive lower endpoint;
- aligned has an upper endpoint for baseline-minus-proposal harm at most .01;
- all journal, as-of, posterior, population and packet-reconstruction checks
  pass, and all scored laws are unchanged by replayed ranking.

Report failures without tuning the weight, prior strength, gate, monitored set
or hidden law on these cohorts. Even a passed screen requires a later actual
serving integration under visible writes, independent provenance and untouched
outcome-labelled agent tasks. Run an independent verifier of source hashes,
Beta updates, scores, packet identities, expected risks and paired screen.
