# Delay diagnostic v86 results

Status: **post-hoc mechanism evidence, not a new confirmation or rescue**.
All 2,560 v85 schedule-runs retain their original predictions, feedback tape
hashes, split decisions and aggregate metrics. The v85 combined-delay failure
is unchanged. These are 640 underlying trajectories, not 2,560 independent ones.

## Artifacts

- [Protocol](mmm-delay-diagnostic-v86-protocol.md),
  [records](mmm-delay-diagnostic-v86.jsonl),
  [summary](mmm-delay-diagnostic-v86-summary.json).
- Diagnostic SHA-256:
  `350f81665c65063bf25591ebcd10a6fab33af0bff82a0ca2a99fd650be1f7403`.
- Parent v85 SHA-256:
  `79e22a74497140af2b75333bcc2701c43a2453181ddf7e377548ec4c60301d0a`.
- The diagnostic pins 123 source/protocol files. The post-hoc evaluator records
  its own hash in the generated summary and independently checks the parent
  records, window totals, source manifest and feedback accounting.

## What the learner was using

The following are confirmation-phase member-shift means in the final window,
steps 448..511. Every cell contains 4,096 forecasts. Brier scores for the inner
predictors use the actual selected observation mask, not an oracle full view.
Weights are the journaled pre-fixed-share weights; forecast smoothing still
applies. The inner subset weight and outer short weight belong to different
levels and must not be interpreted as independent probabilities of correctness.

| Diagnostic | Immediate | Delay 16 | Missing 20% | Jitter + missing |
| --- | ---: | ---: | ---: | ---: |
| Mixture Brier | 0.092989 | 0.179924 | 0.161372 | 0.220257 |
| Inner count Brier | 0.175744 | 0.201969 | 0.216067 | 0.232876 |
| Inner subset Brier | 0.084689 | 0.118758 | 0.133147 | 0.163109 |
| Outer short weight | 91.42% | 38.66% | 60.22% | 17.98% |
| Inner subset weight | 94.09% | 69.12% | 72.23% | 39.75% |
| Subset controls observation | 95.07% | 62.26% | 72.36% | 33.62% |
| Age of newest training example | 34.41 | 49.98 | 42.71 | 52.10 |
| Age of oldest training example | 284.61 | 299.53 | 350.86 | 363.89 |
| Age since model publication | 34.41 | 33.98 | 42.71 | 40.59 |
| Short-window examples from current regime | 75.96% | 70.03% | 59.08% | 54.77% |

A fixed received-label count is not a fixed event-age window. Under combined
stress the 64-label short window reaches roughly 364 steps into the past, well
before the change at step 256. Its oldest-example age is about 79 steps larger
than with immediate feedback. Missing labels also slow the fixed audit-count
refit cadence. The current-regime fraction uses simulator knowledge solely for
diagnosis; the runtime must not receive that label or the true change point.

There is also late selector lag: the subset predictor has lower Brier than the
count predictor, but the inner selector still gives count about 60% weight,
and the outer selector gives the short bundle only about 18%. In this window,
691 of 3,275 delivered member labels are stale for selector updates; they are
not necessarily lost to audited model fitting. These observations do not isolate
the causal contribution of skipping stale updates from ordinary delayed losses.

The analogous common-shift late window has current-regime fraction 53.32%,
subset/count Brier 0.172608/0.236726 and outer short weight 18.46%. The design
phase has the same qualitative pattern. Stationary and null controls remain
included in the summary; no data was selected out to make this story cleaner.

## Why blindly increasing weight is not justified

In combined-stress member steps 384..447, the subset Brier is 0.266770, count
is 0.259429, and neutral is 0.25. The selector's caution in that window is
supported by the actual forecasts. In steps 256..319 the subset is worse still
(0.382883), and only 1.84% of its training examples belong to the new regime.
Prematurely forcing it into control could amplify the old-regime error.

The late-window expert comparisons are descriptive, not a counterfactual gain:
changing weights can change future observation masks and therefore future
training and forecasts. A rescue has to be run end to end on fresh trajectories.

## Next experiments, in order

1. Retain an additional event-age-bounded challenger alongside the existing
   count/subset models and incumbent. Fit only already received, eligible audits
   within a predeclared event-age span, with an explicit minimum support and
   fallback. Keep audit acquisition and initial refit cadence unchanged so the
   first experiment tests age restriction, not extra labels or extra fitting.
   The support/variance tradeoff must be tested on stationary, null, interaction
   and shifted cases; shorter is not automatically better.
2. Separately test a version-aware delayed selector. Retain prediction-time
   lineage and never score a replacement predictor as though it emitted an old
   forecast. Compare delayed-loss handling without the age-window change before
   testing their combination. Missing outcomes remain missing, not imputed truth.
3. Only then test more frequent refitting or optimistic missing-loss hints,
   charging their extra work and keeping hints separate from observed evidence.
   This ordering prevents more compute from disguising a modeling correction.

Relevant primary research: Joulani, Gyorgy and Szepesvari's
[Online Learning under Delayed Feedback (2013)](https://proceedings.mlr.press/v28/joulani13.html)
explicitly associates timestamped feedback with its originating prediction;
its BOLD wrapper updates the originating learner instance. That is a useful
comparison for lineage handling, not a theorem covering our changing model
publications or permanent missingness. Flaspohler et al.'s
[Online Learning with Optimism and Delay (2021)](https://proceedings.mlr.press/v139/flaspohler21a.html)
develops delayed online optimizers with hints for unobserved losses. Its
assumptions and regret analysis would need a separate mapping to our selector;
it does not license treating guessed outcomes as authenticated labels.

## Verification

- Instrumented generation and exact v85 parity: PASS, 169.189 seconds.
- Focused four-schedule parity/race test: PASS, 2.991 seconds.
- Independent summary accounting/source checks and `go vet`: PASS.
- Full diagnostic replay: PASS, 172.466 seconds. Every diagnostic window and
  every original record reproduces exactly against the frozen manifest.

No learner, serving path, OpenClaw, durable store or production setting changed.
No commit or push. All seven research directions remain open.
