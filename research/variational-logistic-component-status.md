# Variational logistic fitter progress

Research-only component; not integrated, not a validated rescue.

Implemented batch Gaussian moments and the optimized likelihood lower bound
from the source-linked variational-logistic-component-plan.md. Every fit starts
from the unit Gaussian prior and xi=1; no previous-window evidence is reused.
The signed ten-feature model retains the MAP comparison's prior and feature set.
The approximate posterior is not an exact Bayesian posterior.

## Numerical audit

The initial 256-iteration cap rejected the 256-identical-row stress fixture.
Increasing the explicit cap to 1024 allowed convergence, but the original
relative fixed-point tolerance 1e-8 left coefficient error 2.20e-7 against an
independent scalar-root solution. Tightening tolerance to 1e-10 passes the
unchanged 2e-7 coefficient and 2e-8 covariance checks. These are component QA
choices, made before fresh predictive quality data, not confirmation tuning.

Tests cover empty-window exact prior, 24 repeated-row configurations with
Sherman-Morrison reference covariance, label complement plus feature and sample
permutation, no input mutation, monotonic bound checks, invalid input and
explicit nonconvergence rejection. Race-enabled focused tests pass (1.541s).
Scoped go vet passes. The iteration cap is not a runtime latency guarantee.

## Predictive and performance checks

The positive-weight sigmoid-Gaussian integral and detached partial-law compiler
are now implemented. Numerical comparison with a wider-domain dense Simpson
integrator passes1521 points (maximum error7.11e-15). All39366 partial-law
checks pass. All1152 consumed generator fits and589824 forecasts pass, with a
maximum56 iterations. The independent artifact integrator has an additional
1449 dense-reference checks. Combined as-of, seed and component race tests pass
in11.974s; scoped vet passes. Numerical integration accuracy is not evidence of
posterior approximation accuracy or calibrated probabilities.

See [frozen contract](variational-logistic-component-contract.md) and
[all benchmark rows](../docs/experiments/mmm-variational-component-benchmarks.json).
On Apple M4,64-label fit costs.180-.201ms, compile.667-.674ms, direct prediction
1.200-1.215us and compiled lookup7.152-7.252ns. The256 repeated-label stress
fit costs27.535-27.757ms and571 iterations. All10 recorded source hashes verify.
These are isolated component timings, not serving latency or quality-matched
costs. Repeated matrix factorization and per-iteration result allocations are
visible optimization opportunities; neither was changed during this comparison.

## Outstanding work

The [v119 full-input experiment](../docs/experiments/mmm-soft-learners-v119-results.md)
retains every v118 case/control and adds explicit same-prior MAP comparisons.
All2688 runs, independent reconstruction and exact replay complete. Neither
variational candidate passes the complete rescue requirements:32-label averaging
passes36/40 MAP gains but harms the four parity-to-majority transitions. Both
pass168/168 MAP non-harm screens under the nonzero.01 tolerance. Consumed
composition remains insufficient; the next distinct lead is segment membership.
Approximate-posterior error, actual observation integration, validated composition
and loaded serving remain unverified. All seven directions stay open.
