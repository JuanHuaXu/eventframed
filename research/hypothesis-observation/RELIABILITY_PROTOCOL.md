# Signal-reliability model averaging v6

Frozen before execution. This is a scoring-only rescue candidate following v5;
acquisition stays with fixed50/50 Joint Gini. All three scoring arms consume the
same16 acquired outcomes and eight signals (cost24). No signal-conditioned
acquisition or stopping changes are included.

## Joint model

Let R be one global episode-level signal mode: reliable, uninformative, reversed,
with prior weights1/3. Let h be uniform over16 hypotheses and m_t independent
Bernoulli(0.5) source modes. Using v4's separately calibrated likelihood matrix
L(signal | m), define L_R as L, uniform0.5, or L with signal columns swapped.

The joint density is proportional to
P(R) P(h) product_t[P(m_t) L_R(s_t | m_t)] times the source-report likelihood
from v3. Condition on all eight signals initially, including their marginal
likelihood in P(R | signals). For each acquired report, update R weights with
each component's BEFORE-update predictive likelihood, then update that
component's Joint posterior. Output the weighted target-class distribution.
This is finite exact model averaging conditional on frozen signal likelihoods,
not integration over calibration uncertainty. Source modes become dependent
after marginalizing R; do not approximate them by independently averaged priors.

Methodological source: Hoeting, Madigan, Raftery and Volinsky (1999),
[Bayesian Model Averaging: A Tutorial](https://sites.stat.washington.edu/www/research/online/hoeting1999.pdf).
The finite provenance model and tests here are ours; the source does not validate
this application or guarantee safety under misspecification.

## Evaluation

Reuse v5's seven generator cases, with new base seed202609150173 and the same
split/case/episode offsets, two splits and128 episodes per case. Calibration
remains frozen. Arms: fixed Joint, reliable-only Joint, model average. Each
episode resets model weights; no hidden modes or target labels update trust.

Confirmation gates, frozen before looking at results:

- Against fixed, mean curve and final Brier harm <=0.01 in all seven cases.
- Against reliable-only on matched_misleading20, mean final gain >=0.03 and
  paired z3.3 descriptive lower bound >0.
- Against reliable-only on independent20, mean curve/final harm <=0.01.

Report all outcomes including random-signal and copied stresses, accuracy,
confidently-wrong fractions and final model weights. These are finite screening
rules; paired normal bounds are not simultaneous confidence guarantees. No
automatic adoption. An inability to distinguish modes is a substantive result.

## Verification

Enumerate all3*16*256 joint states with unequal signal likelihoods and conflicting
observations; check R weights and all target/hypothesis marginals after each
update. Include first-observation invariance of reliability weights: one report
per source group cannot distinguish independent draws from copying. Full replay,
source hashes, normalized probabilities and identical scored observations.
Exclusive-create artifacts, no changes to earlier frozen experiments.
