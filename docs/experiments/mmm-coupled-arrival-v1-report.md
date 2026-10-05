# Arrival selector coupled to acquisition: consumed diagnostic

## Result

Outer-selector acquisition changes do NOT erase the apparent improvement on
these generators. Coupled mean post-change Brier improves over original by
.019-.036 in all eight changed-case/cohort cells. The fixed-view comparison
shows that acquisition adds substantial benefit on reversals but little on some
other cases. This does not overturn v104's failures on different generators or
constitute fresh confirmation. Both original cohorts are consumed.

Second-cohort delayed post-change means (Brier lower is better):

| Case | Original | Arrival, original view | Arrival, coupled view | Original/coupled mean acquisition cost |
| --- | --- | --- | --- | --- |
| Copied bit | .279590 | .260005 | .258504 | 4.554 / 4.636 |
| Copied XOR2 | .279574 | .259701 | .259656 | 4.453 / 4.490 |
| Noisy-copy bit | .280556 | .259742 | .259164 | 4.492 / 4.626 |
| Reversing XOR2 | .273903 | .253957 | .237662 | 4.356 / 4.502 |
| Stable | .047330 | .047604 | .047629 | 4.000 / 4.003 |
| Null | .255241 | .253376 | .253748 | 5.999 / 5.999 |

Cost is full512-frame mean foreground coordinates; Brier is the last256 frames.
Do not combine the differently scoped means into an undocumented utility ratio.
Audits and external monitoring are common and separately retained in raw data.
Changed cases alter approximately74-92 observation masks per trajectory in this
cohort. Stable harm is small but not zero. The first three changed forecasts
remain above the constant-neutral Brier .25, so recovery is still inadequate
in absolute terms. All immediate forecasts agree with original.

## Isolation and verification

New test-only `coupled_arrival_run_test.go`: original conservative controller,
coupled arrival-weighted outer selector, and fixed-view arrival selector. The
coupled arm copies the original INNER selector before each prediction. Thus
this isolates outer-selector/acquisition coupling, not full inner/outer delayed
credit changes. Training audits and fits use the same arrived labels for all
arms; Anti-Pigeon is external and common. Old pre-split credits cannot update
the experimental post-split selector. No label is supplied before arrival.

[Contract](mmm-coupled-arrival-v1-contract.md),
[raw data](mmm-coupled-arrival-v1.jsonl),
[summary](mmm-coupled-arrival-v1-summary.json).
Raw SHA-256:
`60df42ab7121e4f0c4d2cd54cc0b86e82371a6ff711ca10d37a53305f5475f7b`.

192 trajectories/384 paired schedule-runs. Collection checks every original
forecast, input, outcome, arrival, missing flag, audit and original metric against
publication-v1. The independent evaluator reconstructs scores/costs for all arms
and matches the Go fixed-view arm to the previous JS replay.861 captured source
hashes checked; second summary invocation byte-identical. This is not a second
full Go experiment replay. The benchmark file was added after the snapshot.

Targeted race contracts passed2.376s; package vet passed. Collection42.08s.
A duplicate copied type declaration initially prevented compilation; it was
removed before tests or data collection. No data or acceptance criteria changed.
The two journal counters in raw results describe the conservative bookkeeping
objects, NOT the separate arrival selectors' update counts; do not interpret
them as all experimental evidence usage. Arrival eligibility follows recorded
arrival/missing and structural-split boundaries.

## Performance and next action

Three complete512-frame fixture measurements, Apple M4:
immediate141.54/110.49/109.13ms, approximately64.42MB;
delayed93.88/93.88/93.46ms, approximately56.60MB. Includes all three arms,
observations, fits and flush; excludes initial base fitting. Not request latency,
not a candidate-versus-control runtime difference, and not a serving benchmark.

The next discriminating test is cross-generator robustness, including earlier
switch failures, and separating outer-only from inner-credit changes. Do not
repeat an eta/share grid on these consumed results or promote unconditional
role carry. Freeze a broader prospective comparison only after that mechanism
check. All original gates and negative studies stand. All seven goals remain
open; production, whitepaper and remotes unchanged.
