# Sparse variational pilot: numerically verified, unfavorable quality

## Scope and verdict

The frozen consumed index-0 pilot completed 504 fits and 16,128 forecasts:
both phases, all 21 cases, both outcome schedules, clocks 0/128/224, and
64/32-label windows. Every fit retained the original 15 control forecasts.
The collection test passed in 68.84 seconds (69.146 seconds package time).
This includes parsing, fitting, prediction, encoding and output sync; it is
not a serving-latency benchmark or a per-fit microbenchmark.

The candidate is not ready for promotion. All 504 fits exhausted the frozen
64-iteration cap. The integrated forecasts trail strong controls in the
descriptive comparisons below. This is not a full 32-trajectory quality test,
fresh confirmation, or rejection of a converged variational model. No goal
is complete and no production behavior changed.

## Correctness and timing audit

The sparse race suite passed on resumption (9.327 seconds package time),
including as-of poisoning, admitted-label sensitivity, complement symmetry,
Gaussian reconstruction and numerical component checks.

`research/sparse-vrvm-v1-pilot-audit.mjs` independently reconstructs all 504
final Gaussian states with unscaled sample-space Gauss-Jordan inversion and
partial pivoting, distinct from the implementation's scaled Cholesky solve.
It verifies every source origin, excludes unarrived/missing/current labels,
and matches evaluator fields and all 15 controls to the original source tape.
It checks coefficient means/diagonals, Gamma updates, final xi, log determinant,
recorded bound monotonicity, and all 16,128 query moments/probabilities.
Recorded trace monotonicity is not independent recomputation of every ELBO;
the earlier component tests cover the bound's algebra separately.

Integrated probabilities are checked using independent fixed midpoint normal
integration with 32,768 panels on [-10,10], not adaptive Simpson. Maximum
probability discrepancy is 5.19e-10; maximum coefficient-mean discrepancy is
1.19e-10. These finite numerical checks are not a uniform integration-error
certificate. The machine-readable audit is
`mmm-sparse-vrvm-v1-pilot-audit.json`.

## Descriptive forecast results

Expected Brier, lower is better. Each row averages 21 cases times 32 forecasts
at publication clock 224 in phase 1. This is the selected 224-255 window, not
the full terminal 64-frame evaluation. No confidence intervals are claimed.
Immediate and delayed schedules share underlying trajectories; they are not
independent replications.

| Schedule | Label cap | Integrated | Mean plug-in | Same-window generic | Same-window Boolean | Retained Markov |
|---|---:|---:|---:|---:|---:|---:|
| Immediate | 64 | 0.23220 | 0.24798 | 0.15037 | 0.16497 | 0.14291 |
| Immediate | 32 | 0.27191 | 0.30603 | 0.17754 | 0.17200 | 0.14291 |
| Delayed/missing | 64 | 0.21776 | 0.22884 | 0.14825 | 0.16458 | 0.14575 |
| Delayed/missing | 32 | 0.25629 | 0.27873 | 0.17354 | 0.16837 | 0.14575 |

The audit's 17-element Brier arrays are ordered: integrated, plug-in, then
source P[0..14] (generic64, Boolean64, generic32, Boolean32, logistic64,
logistic32, tree64, tree32, variational64, variational32, segment64,
segment32, Markov, static64, static32). All 24 phase/schedule/window/clock
aggregate cells are retained, not just the four displayed rows.

## Next falsifiable lead

The last full-coordinate bound improvement ranges from 0.0385549 to 0.2203065
(median 0.0802771), versus the frozen absolute stopping tolerance 1e-6.
The model is observably not converged under its own criterion. Before changing
priors, feature families, or fitting objectives, isolate convergence with the
same samples, prior, initialization, and equations under a predeclared larger
budget. Charge all extra computation and retain the original capped control.
An improved bound alone cannot establish improved Brier or learning speed.
Do not launch the full trajectory-quality run simply to repeat this weak pilot.

## Artifacts

- Raw SHA-256: `eef26cfc91ed583c0eebc6751d179acbc978b53bf59c32bd8d0f81d3d35c05f1`.
- Audit SHA-256: `6bda95f54fb9588eeb6c498620cae4e729502f3674ff30e805d199f2e8eefb8f`.
- Auditor SHA-256 at the first audit: `31976a994e132e2b7b5173edbd0e711ae130ed5a7abd749553fea9600e0f50d6`.
  It was subsequently extended with an optional convergence-diagnostic mode;
  see the convergence results for the updated source fingerprint.
- Full replay passed in 69.13 seconds (69.456 seconds package time).
  `cmp` confirms byte-identical raw output, hence the same raw SHA-256.

Resumed 2026-09-21 with weekly usage 0% and the user's new above-80% cutoff.
No production, whitepaper, dependency, commit, or push changes.
