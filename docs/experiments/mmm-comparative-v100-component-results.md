# Comparative falsification v100: component checks only

Historical component-stage report. The subsequent
[full experiment](mmm-comparative-v100-results.md) has now run and fails two
positive-recovery gates; the component passes below do not override that result.

Status: COMPONENT PASS; predictive-quality rescue NOT TESTED for this variant.
The full research goal and all seven directions remain open. This does not
replace the v99 result, which fails two positive-recovery gates.

## Mechanism and scope

Implemented `comparative_falsification.go` in the research learner package.
For each target forecast, the alternative mixture puts half its mass on the
neutral forecast and one sixth on each of the three other raw forecasts.
Each alternative mixes 32 predeclared starting times. The rejection boundary
remains 6400 for 64 version/view/expert tests and a conditional 0.01 family
error budget. This conditional guarantee requires the target forecasts to be
the actual conditional outcome laws; it does not certify fitted models true.

Alternative forecasts are captured before the outcome, not treated as
independent votes. Rejections only affect subsequent predictions. Published
versions reset on the fixed 32-step schedule within a 256-step lifetime.
Invalid calls cannot reset a version or mutate its evidence. The raw bank's
cumulative-loss guarantee does not transfer to this gated output.

## Checks actually run

- Literal enumeration of four alternatives and 32 starts matches the recurrence.
- Neutral-only reduction and identical-alternative identities agree numerically.
- One-step conditional expectation holds at extreme full-support probabilities.
- Exhaustive binary histories through depth nine check conditional expectation
  with predictable, history-dependent alternatives at every branch.
- Rejection timing, next-version reentry, missing/wrong/duplicate feedback,
  horizon exhaustion, invalid weights, invalid probabilities, nil receivers and
  issuance atomicity pass.
- Comparative and retained neutral reference/lifecycle tests pass under the race
  detector; `go vet ./internal/observationlearners` passes.

The exhaustive check is a finite numerical test, not a substitute for the
test-martingale argument. Single-owner state is not made concurrency-safe by a
passing race run. No learned-stream run, family-null simulation, delayed-feedback
test, serving integration or production benchmark was performed for v100 here.

## Paired component benchmark

Go 1.27.1, darwin/arm64, Apple M4, default benchmark parallelism 10.
Three 500ms repetitions, no other experiment process observed before starting.
Each operation is an entire 256-step lifetime of four-expert predict/observe
pairs, not one request. Both fixtures use the same fixed forecasts, weights and
outcome pattern; neither measures model fitting, retrieval, persistence or I/O.

| Monitor | ns per 256-step lifetime | Average ns per step | Allocations |
| --- | --- | --- | --- |
| Neutral-only | 56,668-57,345 | 221-224 | 0 |
| Comparative | 282,762-287,445 | 1,105-1,123 | 0 |

The comparative monitor is about five times as expensive in this arithmetic
benchmark. Its absolute cost is small, but no quality gain has yet been measured.
Raw output: [component benchmark](mmm-comparative-v100-component-benchmarks.txt).

## Next falsifier

Freeze the learned-stream and family-null protocol before generating fresh data.
Retain generic, raw-bank and neutral-only paired controls and all 106 quality
gates. Require positive recovery in both directions without sacrificing stationary
protection; do not lower the rejection boundary based on consumed failures.
See [the proposal](../../research/comparative-falsification-proposal.md).
