# Family blend V30: frozen outcome protocol

Frozen 2026-10-03 before design or confirmation collection. All seven whole
goals remain OPEN. This is offline research, not serving or AP authority.

## Design and Sealed Rules

Two cosine-frontier geometries retain n=150 members: sort descending the
two .925 entries plus .65*(cos(step*j)+1)/2+.275, j=1..198, then take150;
step=.005/.020. The declared coordinate is r_i=i/149, a benchmark rank
proxy with no semantic or causal interpretation.

There are32 worlds per geometry/regime,768 worlds per split. Seed base is
2026103003 design and2026103004 confirmation, plus geometry*10000000,
regimeIndex*1000000 and world*1000. Separate truth, label, policy streams
use offsets101,202,303. Common policy streams do not use hidden labels.
No tuning or threshold changes after either split is read.

Regime order and future success probabilities:

1. independent:75 .8 and75 .2 assigned by a seeded uniform permutation;
2. aligned: .9-.8*r;
3. reversed: .1+.8*r;
4. calibrated: b_i;
5. curved: .1+.8*sin(pi*r)^2;
6. shifted_peak: .1+.8*exp(-((r-center)/.14)^2), center=.25+.5*(world%8)/7;
7. alternating: .8 on even indices and .2 on odd;
8. phase_alternating: .2 on even and .8 on odd;
9. permuted_curved: curved probabilities permuted independently of rank;
10. baseline_matched: phi_i~Beta(2*b_i,2*(1-b_i));
11. tree_matched: draw from the26 normalized dyadic tree priors, leaf
    psi~Uniform(0,1), then phi_i~Beta(2*psi,2*(1-psi)); this is conditional
    on the tree family, NOT a draw from the unconditional blend prior;
12. narrow_peak: shifted_peak formula with width .04 instead of .14.

One potential label per member is Bernoulli(p_i). Inference sees only a
selected label after nomination and its pre-label forecast are recorded.
Evaluators alone see probabilities, future/unused labels and tree draws.

## Arms and Matched Controls

Baseline has no evidence. All ten other arms consume exactly32 distinct
labels. Local member Beta, old affine and standalone partition use the same
random-within-stratum design as the primary blend; their selected indices
and labels must agree exactly. Also retain old affine random, uncertainty,
and information policies; blend random, uncertainty and family-information
policies. The blend prior and child models are those of the preflight.
The family-information policy has a .2 uniform exploration floor.

Nomination probabilities are recorded for every arm. Old information and
entropy choices are deterministic given history (probability1); random
draws have probability1/unseen. Strata probabilities are1/unseen-in-bucket,
zero outside that bucket. Blend information has .2/unseen plus .8 at its
deterministic maximizer. This is an active experimental design over a
declared full frontier, not a correction for an unknown selection process.

## Frozen Overall Rescue Gate

Primary arm is blend/stratified_random. Compare it with BOTH local and old
affine stratified_random, receiving identical evidence:

- Every geometry/regime must protect whole-frontier Brier, priority Brier
  and packet usefulness: paired harm upper <=.01 versus each control.
- Curved and shifted_peak must additionally improve whole and priority
  Brier versus old affine by mean>=.01 with lower>0, and usefulness by
  mean>=.02 with lower>0. The narrow and permuted cases remain protection
  controls, not promised nonlinear gains.
- Packet bias abs(mean)+3.5SE <=.10 in every geometry/regime.
- Every arm must have zero duplicate, future-label, posterior/model,
  packet or probability-accounting errors. Isolated model-arm total max
  <=10ms on this cohort; it excludes corpus construction and serving.

Priority weights are3 on the baseline's first10 members and1 elsewhere,
total170. Brier uses (q-p)^2+p*(1-p). Pack the10 highest forecast means,
with original index breaking ties; usefulness is the mean of true p in
that packet. Packet bias is mean(q-p). Intervals are paired mean+/-3.5SE
across32 worlds per cell, NOT simultaneous confidence sequences or
external-law certificates. Overall PASS requires every declared gate in
both splits. No threshold borrowing from a consumed predecessor.

## Observation and Cost Reporting

Compare information versus random and entropy within the same blend, but
these fixed32 observations are a diagnostic, NOT a Goal7 screen at equal
total cost. Report actual construction, selection, update, final-forecast
and total model timings. Allocation microbenchmarks are separate. Random
within-stratum coverage and nonadaptive matched tapes are audited directly.
No result establishes actual-agent outcomes, background freshness, a
continuous-learning convergence theorem, provenance or AP sharing coverage.

The collector exclusive-creates raw files and embeds hashes of model,
tests, collector, both child model sources, independent checker, protocol
and preflight. The checker independently reconstructs flat integrated
likelihoods, family weights, pre-label laws, policy maxima/probabilities,
packets and scores. RNG/Gamma draws are not cross-language regenerated;
an optional Go tape replay supplies an additional generator check.
