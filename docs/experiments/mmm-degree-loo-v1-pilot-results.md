# Degree LOO v1: numerical pilot

## Status

Numerical pilot PASS; broad quality evaluation NOT RUN. All seven goals OPEN.
Do not label training-objective reduction or convergence a prediction rescue.
The pilot is index0 of each phase/case/schedule,84 consumed records, all8
publication clocks, two windows, two LOO objectives, plus linear controls.
It executes4032 fits, of which2688 fit LOO objectives and1344 are linear.

## Mathematics And Verification

The implementation follows the fixed-hyperparameter inverse-diagonal LOO
identity in GPML5.12 and differentiates the two declared losses. It uses
four degree variances with intercept/noise fixed at1. The Gaussian working
model is a mean predictor, not an ordinary Bernoulli posterior. Optimizing
LOO labels is fitting; future as-of forecasts, not that loss, test performance.
See the frozen protocol for source, formulas, caps, and limitations.

| Objective / labels | Converged | Iteration-capped | Mean evaluations |
| --- | ---: | ---: | ---: |
| Smooth squared /64 | 663/672 | 9 | 84.92 |
| Smooth squared /32 | 669/672 | 3 | 78.80 |
| Clipped Brier /64 | 662/672 | 10 | 83.80 |
| Clipped Brier /32 | 667/672 | 5 | 78.38 |

2661/2688 fits meet the declared projected-gradient tolerance. All have finite,
nonincreasing objectives; no line-search caps. None of the27 capped fits is
discarded. Lower fitted losses are not held-out gains. Pilot future labels
are not scored to choose a variant or retune starting points.

Race-enabled identities, explicit leave-one-out refits, held-label flips,
finite differences, invalid inputs, input immutability and model-level
future/missing-label poison tests PASS (2.803s package time). A test initially
failed because its equal-degree parity fixture never saturated, not because
the derivative was wrong. The replacement predominantly-linear majority
fixture exercises65 saturated predictions. The failed log remains alongside
the verified log; only the test fixture changed.

Independent Gauss-Jordan evaluation, not the Go Cholesky path, checks1008
stratified fits,2016 objective values,32256 forecasts and4032 derivatives:
maximum forecast error3.24e-14, objective error4.44e-16 and projected-gradient
error1.77e-11. All2688 optimizer records and source fit-origin lists for the
audited fits are checked. The audit intentionally does not compute future
quality metrics for this single-index pilot.

## Cost

Pilot collection42.88s test-body time; replay43.05s and byte-identical. Full collection has32 times as many
records; that suggests roughly20-25 minutes sequentially, not a measured full
runtime guarantee. Do not silently replace it with pilot-quality conclusions.

Apple M4, Go darwin/arm64,64-label fixture, three300ms benchmark repeats:

| Fit | Time | Allocation |
| --- | ---: | ---: |
| Smooth squared | 38.257-38.636ms | ~57,328bytes /4073allocations |
| Clipped Brier | 41.665-42.025ms | 61,168bytes /4385allocations |

These are isolated fitting costs, not request latency. Prediction uses the
unchanged256-coefficient model; no additional serving benchmark was run.
The new O(4*n^3) derivative is substantially more expensive than marginal
likelihood fitting. It belongs in the slow-path research experiment only.

## Artifacts And Next Action

- `internal/observationlearners/degree_loo_test.go`: evaluator, fitter, tests.
- `internal/observationlearners/degree_loo_run_test.go`: detached as-of adapter.
- `research/degree-loo-v1-audit.mjs`: independent numerical pilot auditor.
- `mmm-degree-loo-v1-pilot.jsonl` and `-pilot-run.txt`: collected results.
- `mmm-degree-loo-v1-pilot-audit.json`: full numerical audit and stop counts.
- `mmm-degree-loo-v1-contracts.txt`: original failed fixture log.
- `mmm-degree-loo-v1-contracts-verified.txt`: corrected race suite.
- `mmm-degree-loo-v1-benchmarks.txt`: measured fit costs.

Next is the full2688-record run under the unchanged protocol, then its840/128
quality gates per candidate, comparisons to old fixed-noise fits, independent
audit and replay. Do not optimize away the negative possibilities or select a
winner from pilot outcomes. No production code, whitepaper, dependency, commit
or push changes were made. The new Go implementation is research-only `_test.go`.

Implementation SHA256:
`8d6b90a1df591316c1918dec4917f458c5b0154146c69c36c9b3eb75c04684c4`.
Adapter SHA256:
`e5266c3c7cf5f00a721831b1ebb60a6b5d6eef04f7201b1e52855f5017aea31c`.
Auditor SHA256:
`ae7506f135ddb59a68b58760d2123f5f28527ba2650d0c260110670cb85f0562`.
Pilot/replay SHA256:
`b1102b5fed2d77825885c8f29168bbf3eb8f4c51b26855b1755fbff36643a096`.
Audit SHA256:
`344a807034797fc1c11d7f152955d47d56616bcb49139691f66dca2ca68dc9d5`.
