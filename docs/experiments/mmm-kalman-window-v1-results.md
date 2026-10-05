# Kalman residual challenger v1: narrow pass, goals open

The frozen [protocol](mmm-kalman-window-v1-protocol.md) passes in both design
and independent confirmation, but it is **not** a Goal 2 or Goal 4 completion.
Its gate emphasizes the last 128 ticks, and therefore misses the candidate's
poor first 64 ticks after abrupt or delayed changes. The nonlinear interaction
control also defeats the fixed linear feature map. Do not integrate into MMM
or production on this evidence.

## Method and audit

The 11-coordinate candidate filters a frozen base model's residual with a
linear-Gaussian working state and Bernoulli labels treated as bounded noisy
observations. It is not an exact Bernoulli posterior. It uses the same arrived
audited labels as the last-64 and loss-window fitted models; the loss-window
detector also uses non-audit arrived labels, matching the existing harness.
Forecasts precede outcome generation and feedback delivery. Delayed labels
are applied at delivery time to origin features, an approximate rather than
exact out-of-sequence update.
All arms receive the full nine-bit context for prediction. This experiment
does not test MMM's selective observation policy or coordinate-acquisition
cost, so its score gains cannot be transferred to that setting without a new
matched study.

- [Design records](mmm-kalman-window-v1-design.jsonl): 160 trajectories,
  SHA256 `0aa7ebbedf798f3fb9a09c9e4ace6e9c3d867e57cbc728c41a70d9877e8a1321`.
- [Confirmation records](mmm-kalman-window-v1-confirmation.jsonl): 160 fresh
  trajectories and distinct base fits, SHA256
  `2121fa4dcffe81e8f676b51cbce29da2ee642acc31280bc1241dcedeb9808adf`.
- The [independent evaluator](../../research/kalman-window-v1-summary.mjs)
  checks source hashes, all 320x512 issued predictions and outcomes, metric
  reconstruction, audit counts, missing/delay schedules, and exact feedback
  accounting. A fresh confirmation replay matched all 161 manifest/record
  entries exactly after excluding runtime timing fields; 97 timing fields
  differed. Focused tests, race tests, `go vet`, and the package's full suite
  pass after the control-parity repair.

## Confirmation outcomes

Arms are frozen base, last-64, adaptive loss-window, and Kalman residual, in
that order. Brier is lower-is-better. The comparison column uses the better
window *within each trajectory*, an optimistic window oracle and therefore a
conservative challenger comparison. Intervals are paired mean +/-3.5 SE over
32 trajectories, not simultaneous coverage or an external-law certificate.

| Case | Tail Brier: base / last-64 / adaptive / Kalman | Tail gain over window oracle | First-64 gain over window oracle |
| --- | --- | ---: | ---: |
| Stable .05 | .06299 / .23711 / .23051 / .06328 | +.16719 | +.17760 |
| Shift at 128 | .37883 / .23705 / .23555 / .05955 | +.17559 [.16417,.18701] | **-.04921** [-.09758,-.00083] |
| Gradual shift | .36697 / .24254 / .24382 / .08924 | +.15212 [.13462,.16962] | +.13963 |
| Delay 16, 25% missing | .37422 / .24516 / .24694 / .09595 | +.14823 [.13277,.16369] | **-.11363** [-.15197,-.07530] |
| Nonlinear interaction | .38571 / .24200 / .24400 / .30308 | **-.06161** [-.07361,-.04960] | **-.12253** |

Stable full-stream Brier is .06224 base versus .06648 Kalman. Mean harm is
.00424 and the paired upper endpoint is .00943, just below the frozen .01
ceiling. That narrow margin does not establish robust non-harm. The `shift128`
and delayed first-64 deficits demonstrate that a late score win does not by
itself give faster recovery. On the gradual case's first 64 ticks, Kalman
Brier .10589 also trails the frozen base's .09446 despite beating the window
models. The interaction loss is the expected falsifier
for a linear residual basis; adding interaction coordinates after seeing this
result would require a new frozen study and compute accounting.

The isolated per-update maximum stream p99 was 709 ns in confirmation, and
the fixed state occupies 1,056 bytes. This measures only the filter update on
one thread, excluding event extraction, fitting, storage, queues, and serving
contention. It is not an EventFrame latency bound.

## Decision

The v1 frozen screen passes, but its late-window gate is too weak to prove
the stated Goal 2 success criterion. This component is a promising cheap
tracker for approximately linear shifts, with a sharp early-adaptation and
nonlinear-model mismatch. Keep it isolated. A successor should predeclare
early-recovery and strong-incumbent gates, use fresh seeds, and compare a
structural/window fallback without suppressing the negative controls. Goal 2,
Goal 4, and all other whole research goals remain open.

Primary method source: [Kalman (1960)](https://www.cs.unc.edu/~welch/kalman/media/pdf/Kalman1960.pdf).
The selective/missing-observation caveat is consistent with
[Sinopoli et al. (2004)](https://eceweb.ucsd.edu/~massimo/Papers_files/Kalman.pdf),
but no theorem from either work is claimed for this Bernoulli working filter.
