# Bounded dependence amplification: mixed metrics, no rescue

Follow the [frozen protocol](mmm-query-amplification-protocol.md). A single
phase0-fitted parameter strengthens pairwise dependence without changing either
marginal, with a factor2 ceiling and a strict margin inside the binary joint
probability boundary. This is research-only, not a production change.

## Results

The fitted alpha is0.536649490882. The training objective is convex; its endpoint
derivatives are -2.023863 and2.331019. Independent reconstruction gives derivative
-4.87e-16 at the fitted alpha, confirming an interior optimum. This alpha is not
a probability or an observed accuracy improvement.

Phase1 delayed forecast diagnostic,672 histories:

| Metric | Independence | Original | Amplified |
| --- | ---: | ---: | ---: |
| Actual joint log loss | 0.993932 | 0.988835 | 0.988368 |
| Target-law expected Brier | 0.170226 | 0.168410 | 0.168634 |

Joint log loss improves slightly while Brier worsens relative to original.
Neither improvement screen passes. These scores average candidate/target pairs
within each trajectory; they are not individual selected-query serving scores.

Actual-publication selection uses the unchanged prediction model and672 paid
queries for each policy:

| Selector | Actual-answer sampled Brier | Teacher-weighted population Brier |
| --- | ---: | ---: |
| Random | 0.167334 | 0.167203 |
| Entropy | 0.166630 | 0.166794 |
| Original joint8 | 0.167366 | 0.166848 |
| Amplified joint8 | 0.167319 | 0.166804 |

Only1/672 phase1 choices changes. Amplification still trails entropy and fails
both selection screens. Across forecast/selection definitions, all four frozen
screens fail. The log-loss gain alone does not rescue the claim or justify
discarding the Brier result.

## Validation and Cost

All2688 records/84cells and287587 query-target pairs are retained. Training uses
143840 pairs from672 phase0 histories; archived missing query outcomes are
explicit offline supervision. Phase1 remains consumed-data evaluation, not
untouched confirmation.

600 unit cases check strict validity, preserved marginals, both label
relabelings and gain scaling; zero motion, exact original endpoint and five
known convex optima pass. Full-data minimum joint-cell probability is0.003218;
admissible caps range1.675044 to2. Pairwise validity is not a full multivariate
process or proof that stronger dependence is universally accurate.

Independent audit uses the joint table's Frechet interval, rather than the
implementation's conditional inequalities, to reconstruct bounds and laws.
9277 acquisition scores,8064 forecast values,26880 publication risks,1092 means
and504 paired bounds pass. It also reconstructs the phase0-only derivatives
from raw labels and verifies the optimum. Replay is byte-identical.

Precomputed-law selection benchmark:2016 calls, median0.008958ms,
p99 0.017834ms, maximum0.523500ms on Nodev26.8.1. This excludes posterior fitting,
retrieval, persistence and serving. It does not establish full acquisition-cost
parity or end-to-end latency success.

Artifacts: `mmm-query-amplification-v1.json`, `-replay.json`, `-audit.json`,
`-benchmark.json`, and `-replay-benchmark.json`; input and script hashes recorded.

## Research Consequence

Global shrinkage, four-cell shrinkage and bounded amplification have not rescued
the acquisition claim. Do not continue adjusting strength bounds or case gates
on these evaluation outcomes. Revisit the joint hypothesis/observation model
and the candidate-observation family, including the unresolved primary-source
misspecification lead, before another policy test. The modest average score
change and valid mathematics do not satisfy any whole-goal completion criterion.
All seven research goals remain open; no production or whitepaper promotion.
