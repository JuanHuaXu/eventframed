# Learned-contrast acquisition v2: frozen component protocol

This Goal 7 study removes the oracle-supplied bit-2 alternative from the
consumed contrast-acquisition v1 study. It tests whether an alternative can be
selected from a small, declared hypothesis family using **arrived labels only**.
It is not an Anti-Pigeon split certificate, automatic hypothesis invention, or
an agent-outcome result.

## Population and separation

- Six 512-tick cases: stable parity, parity-to-bit2, the same shift with
  16-tick label delay and 25% missingness, parity-to-bit0,
  parity-to-parity(0,1,2), and a Bernoulli(1/2) null. Other cases have 20%
  missingness and zero delay. All nonnull rules flip with probability .05.
- Design and confirmation use distinct stream seed bases 2026102201 and
  2026102202. Each split has 16 independently fitted prechange baselines and
  16 independent streams per fit per case (256 paired trajectories/case).
  Fit indices are 400..415 and 500..515. The three policies in a trajectory
  share context, potential outcome, missingness and delay tapes.
- Each policy observes all 512 contexts, nominates exactly 128 of them and
  requests one label per nomination. Missingness and horizon-censored delays
  can reduce *arrived* evidence; report this separately. Current outcomes are
  scored before due labels are delivered. The selector sees only current
  context, the frozen baseline forecast, its own arrived-label state, quota
  and independent selection RNG, never the current or future label.

## Frozen hypothesis family and policies

- The working forecaster has 11 rules: the fitted frozen parity baseline,
  each of the nine single-bit rules, and parity(0,1,2). Alternatives have
  declared 0.05/0.95 Bernoulli probabilities. Initial weights are .90 on the
  baseline and .01 on each alternative. After each arrived requested label,
  recompute weights from at most the latest 32 arrived labels and the fixed
  priors. This rolling-likelihood mixture is a working predictor, not an
  ordinary posterior over the full observation history.
- `random` nominates at the quota rate. `uncertainty` scales that rate by
  normalized frozen-baseline Bernoulli uncertainty. `learned_disagreement`
  names the largest-weight *alternative* using only arrived labels, and
  preferentially nominates current contexts where its predicted class differs
  from the frozen baseline class. For every policy, the remaining quota is
  forced into the final clocks if necessary. All three use the **same**
  rolling-likelihood forecast model, updated only by their own arrived labels.
- The alternative family contains the three changed rules by design; the
  experiment tests evidence-driven *selection among known rules*, not a new
  representation learner. The parity(0,1,2) case is a nonlinear stressor
  relative to single-bit alternatives, not a universal out-of-family test.

## Frozen screen

Measure proper Brier on every issued prediction: full512, first64 after the
change, and post256. Use paired mean +/- 3.5 standard errors over 256 streams
per case as a descriptive gate, not simultaneous population coverage. Retain
every case, score and failure. A component pass requires **both** splits:

1. In immediate and delayed bit2 shifts, post256 Brier gain over **each**
   random and uncertainty control is at least .01, with a positive paired
   lower endpoint. First64 mean harm to either control is at most .005.
2. Stable and null full512 Brier harm upper endpoint versus each control is
   below .01. Bit0 and parity(0,1,2) post256 harm upper endpoint is below .01.
3. At least 60% of postchange nominations in both bit2 cases and the bit0
   case name the correct alternative. Arrived-label mean gaps against either
   control are at most two per case, and all arms request exactly 128 labels.

This strict joint screen may fail even if one learned contrast improves. It
does not confer Anti-Pigeon authority, target-law guarantees, or serving
latency claims. A failure cannot be rescued by threshold tuning on these
cohorts. Method motivation: hypothesis-discriminating acquisition in
[Golovin, Krause and Ray](https://arxiv.org/abs/1010.3091), without claiming
their algorithm or guarantees.
