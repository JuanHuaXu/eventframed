# Rank-Certainty Modulation: Codex Replay Regression

Date: 2026-08-29

This full replay uses the same frozen Codex-session cutoff, dual calibration
parameters, OpenAI-compatible Nomic embeddings, and isolated LibraVDB contract
used by `calibration-rescue-real-rerun-2026-08-29`. The changed implementation
uses rank-boundary certainty for plasticity and an independent accepted-path
reliability gate. Because this split has already been inspected, these are
regression results, not a new untouched confirmation claim.

## Results

| Block and metric | Raw baseline | Previous full upgrade | Rank-certainty upgrade | Change vs previous |
| --- | ---: | ---: | ---: | ---: |
| Design Recall@10 | 40.11% | 47.69% | 56.74% | +9.05 pp |
| Design packed recall | 27.87% | 33.53% | 41.07% | +7.54 pp |
| Design MRR | 72.08% | 80.06% | 87.52% | +7.46 pp |
| Confirmation Recall@10 | 39.44% | 42.85% | 46.50% | +3.65 pp |
| Confirmation packed recall | 28.56% | 32.56% | 35.79% | +3.23 pp |
| Confirmation MRR | 84.53% | 89.88% | 93.27% | +3.39 pp |

The confirmation comparison against the raw baseline reports a +7.06 percentage
point Recall@10 gain with cluster-bootstrap 95% interval [5.05, 22.49] and a
+7.23 point packed-recall gain with interval [1.77, 22.49]. Mean packed token use
fell by 10.26 tokens and high-priority misses did not increase.

## Calibration boundary

The forecast law did not change, as intended. Confirmation Brier is exactly the
previous `0.2881675856`; design Brier is equal up to floating-point noise. The
confirmation Brier gain over baseline remains inconclusive, and the design block
still regresses from `0.2056849962` to `0.4744088777`. Rank-certainty modulation
therefore rescues ranking and packing behavior; it does not rescue the separate
cross-regime probability-calibration problem.

## Interpretation

Using calibrated usefulness as both prediction and plasticity control was the
wrong coupling. A local packing-boundary gap answers the operational question
"how settled is this top 10?" without pretending to estimate truth. The accepted
Bayesian, residual, or graph path separately answers "is this correction allowed
to act?" The replay supports that separation for ranking, while retaining the
hard delta cap and failing closed when no correction path is accepted.

The isolated run completed 1,286 design cases and 138 confirmation cases. Full
datasets, variant reports, paired comparisons, and protocol metadata are stored
beside this file.
