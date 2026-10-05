# Full-vector population correction: modest additional rescue

The [frozen experiment](VECTOR_ALLOCATION_PROTOCOL.md) removes the fixed-line
restriction and optimizes all four outcome probabilities at each history.
It retains the same average genuine-risk objective, fixed-.20 baseline,176
population coefficient constraints over noise[.10,.30], and all900 gates.

Result:889/900 (FAIL). All800 population protection and50 false-confidence
checks pass. Eleven genuine-gain requirements remain unmet: three at .25 noise,
eight at .30. The line-restricted average-objective rule passed888/900.

| Actual noise | Pass /180 | Worst population harm | Genuine gain range |
| --- | ---: | ---: | ---: |
| .10 | 180 | .001116 | .011631 to .013023 |
| .15 | 180 | .005746 | .009474 to .011701 |
| .20 | 180 | .008503 | .005980 to .010320 |
| .25 | 177 | .009755 | .001864 to .008724 |
| .30 | 172 | .010000 | -.002797 to .007873 |

Average-objective Brier improves over the line-restricted optimum in all ten
allocations by .000746-.001627. That is a genuine change in the optimized
quantity, not uniform improvement across worlds. Some noisy genuine cases still
worsen. More freedom plus an average objective does not ensure the minimum
gain holds in each environment.

## Solver and verification

The separable Lagrangian minimizer is a normalized nonnegative combination of
joint class masses. It is therefore a valid probability vector for each
history. All ten solves converge in91-1069 sweeps, within the unchanged5000 cap,
with numerical primal-dual gaps below1e-8 and risk<=.01+1e-12.
The solution is a constrained decision forecast, not ordinary Bayesian
authentication of source claims.

Full optimization replay is byte-exact. Original controls remain identical
within their declared comparison tolerance in all worlds. Alternate Brier
scoring with fixed stored forecasts agrees on3200 values within7.77e-16;
all gate decisions agree. Eight hundred direct/population-polynomial
comparisons agree within2.06e-14.

Four analytic binary cases, twelve two-history feasible-grid comparisons and
two invalid-input checks pass. These checks verify numerical consistency,
not an independently implemented full optimizer or a formal roundoff proof.

[Full forecasts and results](vector-allocation-experiment.json),
[verification](vector-allocation-verification.json),
[solver](vector-risk-allocation.mjs).

## Tradeoff and next step

Worst SINGLE history/outcome loss increase remains .334035, while population
harm is bounded by .01 under the assumed family. No per-event or high-priority
safety guarantee follows. A real deployment would need separately specified
local protections and evidence that its generating law is covered.

The prior fixed-line impossibility bounds no longer apply to this larger
decision family. To distinguish objective imbalance from information limits,
test endpoint or worst-regime optimization in this full-vector family and
retain numerical dual bounds. Do not lower gain thresholds or discard noisy
worlds. The .30 endpoint is not presumed to be the worst regime for every rule.

This is exact finite model-based design on consumed cases, not new empirical
confirmation, actual-agent acquisition-speed validation or a scalable serving
implementation. No timing benchmark, production change, whitepaper promotion
or publication. All seven whole research directions remain open.

