# Consumed-v10 ridge specialist screen

2026-10-01. Outcome: **not a general rescue**. This is a component screen on
already consumed v10 data, not fresh confirmation or an adopted MMM policy.
The [frozen protocol](mmm-ridge-v10-component-protocol.md) predates this run.

The opt-in research runner replayed both complete v10 journals: 480 streams per
input mode, 10 scenarios, two original phases, three fitting groups and eight
streams per group. It independently regenerated all inputs/outcomes and rejected
any mismatched outcome, selected partial view or delivered-label clock. Ridge
used the last 64 *delivered paid-audit* labels, refit from 32 audits at every
16th audit, and forecast before incorporating deliveries at that clock. It
used fixed uniform completion weights in both uniform and latent-correlated
input modes. Before its first successful fit it returned the archived incumbent.
All 6,062 fit/compile attempts succeeded; failures would have remained in the
denominator and retained the prior published law.

The independent JS checker recomputed scores from the complete forecast tape,
checked source hashes, all 960 stream identities, publication ordering and
same-clock exclusion, and fallback equality. The table reports confirmation-
phase post-window Brier for arm 1 (the replacement MMM observer); lower is
better. Gain is incumbent minus ridge, averaged over 24 trajectories per cell.
The 3.6-SE lower bound is a diagnostic reused from v8, **not** a new
confirmation interval.

| Input | Scenario | Incumbent | Ridge | Gain | 3.6-SE lower |
|---|---|---:|---:|---:|---:|
| Uniform | stable05 | 0.04820 | 0.06674 | -0.01854 | -0.02596 |
| Uniform | stable20 | 0.16400 | 0.19642 | -0.03242 | -0.04145 |
| Uniform | shift128 | 0.24625 | 0.23435 | +0.01190 | -0.00137 |
| Uniform | shift256 | 0.25732 | 0.27333 | -0.01601 | -0.03474 |
| Uniform | interaction | 0.21536 | 0.19283 | +0.02253 | +0.00869 |
| Uniform | null | 0.25184 | 0.27872 | -0.02688 | -0.03434 |
| Latent | stable05 | 0.04846 | 0.07694 | -0.02847 | -0.03544 |
| Latent | stable20 | 0.16177 | 0.19436 | -0.03259 | -0.04387 |
| Latent | shift128 | 0.24531 | 0.23450 | +0.01081 | -0.00149 |
| Latent | shift256 | 0.25621 | 0.29157 | -0.03536 | -0.05915 |
| Latent | interaction | 0.18926 | 0.17314 | +0.01612 | +0.00553 |
| Latent | null | 0.25200 | 0.27634 | -0.02434 | -0.03192 |

This specialist has useful capacity on the interaction case, but fails the
unchanged primary shift-256 recovery and stable/null protection requirements.
It also harms late shift, recurring and delayed/missing cases in the full
cell summaries. The positive shift-128 means do not clear the 3.6-SE bound.
The interaction result is exploratory after looking at the same consumed
family, not a license to add a scenario-aware router.

Focused `go vet` and `go test -race` passed. A full uniform replay under the
race detector also passed (97.70 seconds); excluding timing fields, its
prediction and publication tape had the same SHA-256 as the normal replay:
`b1a4e804e5d16793afbd4006bac73c2b4adffb4df5215cb52fb961f7ae9eff22`.

Cost, measured inside this single-threaded offline replay: each mode had
3,031 successful publications, 181,462 ready frames, about 129-132 ms total
fit time and 181 ms total conditional-table compilation time. Forecast lookup
time summed to about 10-11 ms across four arms on ready frames. Whole replay
wall time was about 4.1 seconds per mode, including journal decode and score
capture. These are *component* timings, not service p95/p99 or concurrent
memory/latency evidence; the compiled ternary law has 19,683 cells.

Artifacts: [uniform tape](mmm-ridge-v10-uniform.json.gz),
[uniform independent scores](mmm-ridge-v10-uniform-check.json),
[latent tape](mmm-ridge-v10-latent.json.gz), and
[latent independent scores](mmm-ridge-v10-latent-check.json). The runner is
`internal/observationlearners/research_ridge_v10_run_test.go`; independent
checker is `research/check-ridge-v10.mjs`. Each tape records journal,
protocol and source SHA-256. No production serving path, OpenClaw instance,
whitepaper or remote was changed.

Next: do not adopt raw ridge or tune a gate on this consumed corpus. For goals
2 and 4, a new learner needs a predeclared stationary/null safeguard and a
separately generated outcome family; for goal 5, prospective outcome-labeled
agent tasks remain the stronger discriminating test. The interaction gain may
motivate a dedicated, *freshly tested* specialist, but is not general evidence
of robust recovery.
