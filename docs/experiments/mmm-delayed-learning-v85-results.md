# Delayed/missing retained learning v85 results

Verdict: **FAIL the frozen overall adoption rule in both phases.** Immediate,
fixed-delay and independent-missing schedules retain relative gains, but the
combined jitter/missing schedule fails both primary shifted cases. This does
not reverse the earlier immediate-feedback results or establish robust learning.

## Design and provenance

- [Frozen protocol](mmm-delayed-learning-v85-protocol.md),
  [raw records](mmm-delayed-learning-v85.jsonl),
  [machine-readable summary](mmm-delayed-learning-v85-summary.json).
- SHA-256: `79e22a74497140af2b75333bcc2701c43a2453181ddf7e377548ec4c60301d0a`.
  The artifact pins 119 source/protocol/evaluator files.
- 640 fresh underlying trajectories, each reused under four feedback schedules:
  2,560 schedule-runs. Each phase/scenario/schedule has 64 trajectories, with
  512 forecasts per arm. These are not 2,560 independent underlying streams.
- Candidate and count control share all evidence acquisition, delivered labels,
  publication times and authorized split transitions. Forecasts are issued
  before simulator truth and learning only consumes delivered packets.
- Post windows are steps 256..511, except recurring steps 128..511. Intervals
  below use paired trajectory means and the frozen z=3.5 screen; they are not
  exact finite-sample or time-uniform confidence certificates. Results remain
  conditional on the fitted bases and these synthetic generators.

## Confirmation results

Brier is lower-is-better. Gain is count-control Brier minus candidate Brier.
Each member/common cell requires mean gain >=0.005 and a positive lower bound.

| Scenario | Feedback | Count Brier | Candidate Brier | Gain interval | Gate |
| --- | --- | ---: | ---: | --- | --- |
| Member shift | Immediate | 0.240308 | 0.204891 | 0.035417 [0.029208, 0.041626] | PASS |
| Member shift | Fixed delay 16 | 0.266249 | 0.255183 | 0.011066 [0.005836, 0.016296] | PASS |
| Member shift | Missing 20% | 0.250574 | 0.234017 | 0.016557 [0.010109, 0.023006] | PASS |
| Member shift | Jitter 0..31 + missing 20% | 0.266478 | 0.264536 | 0.001942 [-0.001403, 0.005286] | FAIL |
| Common shift | Immediate | 0.242384 | 0.205470 | 0.036915 [0.030050, 0.043779] | PASS |
| Common shift | Fixed delay 16 | 0.263733 | 0.249351 | 0.014382 [0.009036, 0.019728] | PASS |
| Common shift | Missing 20% | 0.249240 | 0.234282 | 0.014958 [0.008824, 0.021092] | PASS |
| Common shift | Jitter 0..31 + missing 20% | 0.267842 | 0.265487 | 0.002355 [-0.001252, 0.005962] | FAIL |

Design-phase combined gains also fail: member 0.001449
[-0.001889, 0.004788], common 0.002313 [-0.001395, 0.006021]. All other
36 of the 40 phase/scenario/schedule cells pass their predeclared rules. The
stable/null/recurring rule is non-harm versus the equally delayed count control,
not absolute recovery or an improvement requirement.

| Confirmation candidate | Immediate | Delay 16 | Missing 20% | Jitter + missing |
| --- | ---: | ---: | ---: | ---: |
| Member-shift accuracy | 66.68% | 59.61% | 60.86% | 56.32% |
| Common-shift accuracy | 66.80% | 60.78% | 60.24% | 56.11% |
| Recurring accuracy | 74.39% | 69.45% | 72.54% | 67.44% |
| Stationary accuracy | 94.80% | 94.80% | 94.81% | 94.80% |

Absolute degradation matters: candidate member Brier rises by 0.059645
[0.052306, 0.066984] under combined stress versus its paired immediate run.
Both stressed member/common candidate Briers exceed the 0.25 score of a
constant 0.5 forecast. Thus a relative pass in some schedules is not proof of
adequate absolute probabilistic performance. Stationary protection is retained;
it does not make the shifted predictor 94.8% accurate.

## Feedback and bounds

Counts below cover all 320 confirmation trajectories per schedule, for either
arm separately. Every schedule has 163,840 issued forecasts. Censored includes
missing labels and packets not yet due at the fixed experiment endpoint.

| Schedule | Selector-applied | Stale selector feedback | Censored | Maximum pending |
| --- | ---: | ---: | ---: | ---: |
| Immediate | 163840 | 0 | 0 | 1 |
| Delay 16 | 125487 | 33233 | 5120 | 17 |
| Missing 20% | 131244 | 0 | 32596 | 22 |
| Jitter + missing | 107898 | 19389 | 36553 | 35 |

No labels are invented. Received stale labels may still enter eligible audited
training once; stale counts mean skipped current-selector updates, not deletion
of training evidence. Member-shift combined runs average 100.17 received audits,
4.73 subset fits and 67.75 stale selector updates, compared with 128.89 audits,
6.58 fits and zero stale updates for immediate feedback.

Foreground acquisition stays within six coordinates per forecast. The shared
monitor costs eight additional coordinates per step; shared audit cost averages
about 4.5 more. These costs are not hidden inside the foreground cap. The
64-entry journal bound holds. No new performance claim is made: this study
changes the experiment driver, not the frozen v84 journal or v82 learner.
Refitting occurs synchronously between logical simulation steps, so this is not
a concurrent-serving or durable-persistence latency benchmark.

## Verification and interpretation

- Fresh generation: PASS, 169.435 seconds including test-package completion.
- Focused race tests: immediate journal parity, lifecycle and delayed-learning
  contracts PASS, 3.367 seconds. All-missing feedback produces no fitting,
  selector learning or split; schedule changes preserve the latent-data hash.
- Evaluator validates source hashes, seed/order, paired latent trajectories,
  metrics, equal policies and full accounting. `go vet` on observationgate,
  observationlearners and observationpreserved passes; `git diff --check` passes.
- Full source-manifest replay: PASS, 172.335 seconds. All 2,560 records,
  prediction/feedback tape hashes and 640 paired latent trajectories reproduce
  exactly against the 119-file frozen manifest.

The failure is confirmed; its cause is not isolated. Three plausible contributors
are increased event-age span in the received-label window, delayed refit support,
and selector feedback invalidated by publication/split versions. Lowering the
acceptance threshold, treating missing labels as negatives, or applying stale
feedback to unrelated model versions would not be a valid rescue.

Next investigate those contributors with read-only age/support/weight diagnostics
on these frozen tapes, then freeze a fresh matched experiment for an age-aware
challenger or publication policy. Do not present post-hoc diagnostics as new
confirmation. All seven research directions remain open; no production,
OpenClaw, database, commit or push changes were made.
