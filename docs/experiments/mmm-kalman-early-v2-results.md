# Kalman early-recovery v2: process-noise rescue fails

The [frozen component protocol](mmm-kalman-early-v2-protocol.md) **fails** in
both independent design and confirmation cohorts. Increasing random-walk
process noise repairs part of the v1 abrupt-change early deficit, but does not
repair delayed early recovery and harms stationary and nonlinear controls.
Do not integrate either variant into MMM or production. Goals 2 and 4 remain
open, as do the other five whole research goals.

## Confirmation result

Brier is lower-is-better. Each cell is a mean over 32 paired trajectories,
four independently refitted baselines times eight streams. Every arm received
the same context, audit, missingness, outcome and delivery tape. `q=0.0005`
is the original Kalman working filter; the alternatives were frozen at
`q=0.002` and `q=0.01` before either new cohort was read.

| Case and segment | Frozen base | Last-64 | q=.0005 | q=.002 | q=.01 |
| --- | ---: | ---: | ---: | ---: | ---: |
| Shift128, first64 after change | .38757 | .25163 | .30073 | .26847 | **.23935** |
| Delay16 + 25% missing, first64 | .37564 | **.25153** | .37013 | .35485 | .34324 |
| Gradual, last128 | .37514 | .24486 | .09291 | **.07721** | .08079 |
| Stable, full512 | **.06318** | .24212 | .06971 | .07418 | .08492 |
| Interaction, last128 | .38409 | **.24389** | .29780 | .32519 | .36082 |

The `q=.01` early abrupt-shift gain over the original filter is .06138
(paired mean-minus-3.5-SE lower endpoint .04221), and it beats last-64 by
.01229 in mean. The same variant still loses .09172 Brier to last-64 in the
delayed early segment. `q=.002` loses .10333 there. Neither meets the frozen
early gate. Stationary full-stream harm versus the frozen base is .01099
(upper endpoint .01859) for `q=.002` and .02174 (upper .03068) for `q=.01`,
both above the .01 ceiling. Their nonlinear last128 harm versus last-64 is
.08131 and .11693. Both meet the late linear-shift and component cost gates,
but the overall screen is **FAIL** in each split.

Design agrees on the failure pattern: delayed first64 Brier .34310/.33361
for the two faster filters versus .25137 last-64; stable full harm .00794
(upper .01374)/.01728 (upper .02461); nonlinear last128 harm .08921/.12249.
The maximum measured per-stream filter-update p99 was below 50 us for each
candidate, and state size was 1,056 bytes. This times only `Observe`, not
extraction, filtering of the stream, refits, queues, persistence or serving.

## Audit and limits

- [Design records](mmm-kalman-early-v2-design.jsonl): 160 trajectories,
  SHA256 `fc7c60c6ed31326c126eba9bc8ddec39811b162ffeb37ae5eaaa9d6323ccb00b`.
- [Confirmation records](mmm-kalman-early-v2-confirmation.jsonl): 160
  trajectories, SHA256 `fe168278174624879307d2cbbe43e96ef4775da9eb56e178fa49bc51b2b3de1b`.
- The [independent verifier](../../research/kalman-early-v2-verify.mjs)
  checks source hashes, unique rows, all 163,840 forecast ticks (819,200
  probabilities) and scores, no future delivery, audit/update counts, pending labels, paired
  intervals and each frozen decision gate. A fresh confirmation recollection
  matched all 161 JSONL entries exactly after excluding per-update timing.
  Focused race tests and `go vet` pass.

The filter remains a linear-Gaussian working approximation to Bernoulli
residuals. Delayed origin features are applied when labels arrive, not through
an exact out-of-sequence smoother. The nonlinear interaction is outside its
fixed feature span. The first 16 post-change delayed clocks cannot contain a
post-change audited label under this schedule; process noise alone cannot
identify the new rule before relevant evidence arrives. A future study should
test a distinct onset-aware, delayed-evidence or nonlinear challenger on fresh
seeds, retaining strong incumbents and the stationary false-alarm guard.
