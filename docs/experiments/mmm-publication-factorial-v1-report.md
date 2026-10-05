# Publication failure: withholding and reset have different costs

Consumed-data diagnosis, not a fresh confirmation or a successful rescue.
The original publication-v1 failure stands. Both factor changes are harmful in
important cases, but their relative effects depend on the feedback schedule.

## Integrity

All192 original trajectories, both schedules and four factorial arms were run.
The unchanged and holdout-reset endpoints match the parent in every forecast,
score, split time and latent tape. Complete Go replay matches all384 runs,
including the two new arms. Race-enabled endpoint contracts passed in3.213s;
package vet passed. Collection78.36s; full replay78.95s.

- [Predeclared diagnostic contract](mmm-publication-factorial-v1-contract.md)
- [Raw results and source snapshots](mmm-publication-factorial-v1.jsonl)
- [Independent summary](mmm-publication-factorial-v1-summary.json)
- [Byte-identical summary replay](mmm-publication-factorial-v1-summary-replay.json)
- [Evaluator](../../research/publication-factorial-summary.mjs)

Raw SHA-256: `1533a049c2146cc7839799e9a394691d96c3fdf4a1a8e8d1c504db86053bf8e6`.
Parent SHA-256: `566710fe18c30d2fa905b4c5a9203874bf25d965b73da846c15cff2efcddc596`.
All843 captured source snapshots match. The replay/benchmark source was added
after that snapshot. The evaluator independently reconstructs scores,
availability, partitions, generation accounting, reset weights and parent
endpoints. Both algebraic decompositions of endpoint harm agree.

## Factorial results

Second-cohort post-change Brier, lower is better:

| Case | Schedule | All/carry | All/reset | Holdout/carry | Holdout/reset |
| --- | --- | ---: | ---: | ---: | ---: |
| Copied bit | Immediate | .194749 | .199823 | .236588 | .239361 |
| Copied XOR2 | Immediate | .212278 | .215999 | .243304 | .249193 |
| Noisy-copy bit | Immediate | .201917 | .206682 | .239364 | .241968 |
| Reversing XOR2 | Immediate | .180795 | .187420 | .225035 | .226834 |
| Copied bit | Delayed | .279590 | .305960 | .289600 | .319943 |
| Copied XOR2 | Delayed | .279574 | .310967 | .280094 | .319049 |
| Noisy-copy bit | Delayed | .280556 | .319280 | .287676 | .329839 |
| Reversing XOR2 | Delayed | .273903 | .305432 | .280724 | .319326 |

With immediate feedback, withholding while preserving weights adds.0310-.0442
Brier across these four cases. Resetting with all training labels adds only
.0037-.0066, with three of those four approximate intervals crossing zero.

Under delay, resetting while retaining all training data adds.0264-.0387 Brier,
with all four approximate intervals above zero. Withholding while retaining
weights adds.0005-.0100. For delayed copied-XOR2, the effects interact: the reset
penalty grows from.03139 to.03895 after withholding, interaction.00756 with
interval[.000776,.014348]. Do not add marginal effects as if independent.

Stable cases also penalize resets: delayed stable Brier .047330 rises to.050057
with all-data reset, while holdout/carry stays.047332. Null cases and both
original cohorts remain in the summary. No adoption gate was invented for this
consumed diagnostic. Reported intervals are descriptive mean +/-3.5SE over16
trajectories, not fresh confirmation or research-wide error control.

These are total effects within the finite simulation. Each arm chooses its own
observations, so acquisition changes are included. The experiment does not prove
that either mechanism explains all errors in arbitrary event streams.

## Could a fixed blend rescue the calibrated branch?

A separate [analytic diagnostic](../../research/publication-mixture-headroom.mjs)
uses the original unchanged and warm-calibrated forecasts from publication-v1.
For p=unchanged, q=warm and observed y, define d=q-p. The empirically best
constant convex weight is:

```text
alpha* = clip(-sum((p-y)*d)/sum(d*d), 0, 1)
gain = [-2 alpha* sum((p-y)*d) - alpha*^2 sum(d*d)] / n
```

When the denominator is zero choose alpha=0. Literal score reconstruction,
identical forecasts, endpoint optimum and interior optimum checks pass. A local
self-test initially distinguished IEEE signed zeros; its assertion was corrected
before producing any output artifact. No experiment or threshold changed.

[Per-trajectory headroom](mmm-publication-v1-mixture-headroom.json) chooses the
best constant weight separately on each trajectory's scored post-change labels.
This is deliberately optimistic and not deployable. Even this bound misses
.005 mean gain in5/8 delayed changed-case cells. Second-cohort maxima are
.00453 copied bit, .00476 copied-XOR2, .00355 noisy-copy bit, .00944 reversal.
It ignores extra observation/computation cost. It rules out a sufficient static
blend on those consumed forecast tapes, not time-varying policies, future data,
or a new model with different forecasts.

## Decision

Preserve incumbent training freshness and accumulated selection evidence.
Do not rescue this design by changing only the holdout size, using a permanent
blend, or interpreting calibration's wins against a weakened control as progress
over the original. Any further publication-selection method must avoid both
costs or justify them with stronger forward evidence. Next research should
examine validation without excluding the newest data and explicit transfer
error when proxy validation scores are assigned to a full-data model.
That is a research direction, not an implemented or validated rescue.

## Compute

Apple M4 darwin/arm64, complete four-arm512-frame fixtures, three1x repetitions:
immediate265.834 /238.467 /241.266ms, about175.858MB and60244-60249 allocations;
delayed178.744 /176.788 /179.404ms, about132.528MB and56719-56725 allocations.
Command: `go test ./internal/observationgate -run '^$' -bench '^BenchmarkPublicationFactorialFixture$' -benchtime=1x -count=3`.
These are diagnostic whole-fixture costs, not request latency or a production
implementation. All measurements are retained. No production, whitepaper,
commits or remote changes occurred. All seven research goals remain open.
