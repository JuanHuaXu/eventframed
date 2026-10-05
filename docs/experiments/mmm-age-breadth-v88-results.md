# Age challenger breadth v88 results

Verdict: **FAIL overall breadth adoption in both phases**. The combined-delay
four-bit parity case fails the frozen improvement rule. The other six changed
cases pass in both phases, as do the stationary/null and immediate-feedback
protection checks:38 of40 cells pass, but the required conjunction does not.
The matched v87 result remains valid within its narrower scope.

## Evidence

- [Protocol](mmm-age-breadth-v88-protocol.md),
  [records](mmm-age-breadth-v88.jsonl),
  [summary](mmm-age-breadth-v88-summary.json).
- SHA-256:
  `4d97446e78948d8544c218c9e195e8d446ea5cb086c63426a7c79c54a5e14dc0`.
  The artifact pins128 source/protocol/evaluator files.
- 1,280 underlying fit/stream pairs, each run under immediate and combined
  jitter0..31/missing0.2 feedback:2,560 schedule-runs. Each trajectory has a
  separately fitted4096-label base, reused across its two schedules. This adds
  fitting-sample variation rather than conditioning on one favorable base.
- Ten generator cases, two phases,64 trajectories per cell,512 forecasts per
  arm. Post metrics cover steps256..511. All parameters of the v87 candidate
  remain unchanged. The uniform-input assumption remains misspecified on the
  declared dependent-input cases; it was not repaired after seeing these data.
- No extra audits, labels, observer budget, selector update rule or fit trigger.
  The retained learner remains the control. Extra age fitting remains charged.

## Confirmation under combined stress

| Case | Retained Brier | Age Brier | Retained accuracy | Age accuracy | Primary result |
| --- | ---: | ---: | ---: | ---: | --- |
| Single bit | 0.266449 | 0.214223 | 56.48% | 69.65% | PASS |
| XOR2 | 0.268145 | 0.225845 | 55.69% | 67.46% | PASS |
| Majority3 | 0.267686 | 0.249147 | 55.38% | 62.79% | PASS |
| Multiplexer | 0.270126 | 0.250047 | 54.92% | 62.52% | PASS |
| Parity4 | 0.275196 | 0.271723 | 51.68% | 53.91% | FAIL |
| Dependent single bit | 0.254873 | 0.198431 | 59.53% | 72.60% | PASS |
| Dependent XOR2 | 0.262465 | 0.214935 | 57.74% | 69.65% | PASS |

Primary rules require mean post-Brier gain >=0.005 and positive paired z3.5
lower bound. Confirmation gains and intervals:

| Case | Mean gain | Paired interval |
| --- | ---: | --- |
| Single bit | 0.052227 | [0.040187, 0.064267] |
| XOR2 | 0.042300 | [0.032919, 0.051681] |
| Majority3 | 0.018539 | [0.010983, 0.026096] |
| Multiplexer | 0.020079 | [0.013311, 0.026847] |
| Parity4 | 0.003473 | [-0.000682, 0.007628] |
| Dependent single bit | 0.056442 | [0.046985, 0.065899] |
| Dependent XOR2 | 0.047530 | [0.037755, 0.057306] |

Design parity also fails: gain0.002941, interval[-0.000663,0.006544]. The
failure is not caused by a single confirmation threshold crossing. Intervals
are the predeclared approximate trajectory screen, not exact simultaneous
coverage, confidence sequences or calibration certificates.

Absolute quality limits the interpretation. Parity Brier is worse than the
constant0.5 score of0.25, and accuracy remains near chance. Multiplexer Brier
is slightly worse than0.25 and majority is only slightly better; their relative
passes do not establish strong probabilistic performance. Under immediate
feedback parity improves0.241587 to0.224785, accuracy56.84% to62.38%, still
far below stationary performance.

Stationary combined accuracies are94.99% (uniform) and95.03% (dependent), each
unchanged between arms. Null Brier is0.252254 to0.252428, with near-chance
accuracy. Its protection rule passes, but there is no claimed null skill.

## Interpretation and next falsifier

The age rescue generalizes to some interactions and to the two tested
dependency mappings across independently fitted bases. It does not solve all
higher-order interactions. Its remaining parity failure could arise from:

1. The small recent sample failing to learn a useful joint predictor.
2. Observation selection failing to acquire the joint variables it learned.
3. Selector lag keeping an improved predictor from controlling the forecast.

The observer's source shows one-step expected-information-gain view selection.
For uniform parity, an unobserved independent required bit makes the true
conditional outcome probability0.5, so some individually evaluated views can
have zero immediate information despite useful combinations. This is a possible
mechanism, not proof of an implementation error: the existing view order and
tie-breaking can still acquire a sufficient combination within the budget.

Next perform a read-only parity/XOR/bit diagnostic on these frozen runs:
compare each available model's actual-mask and full-input predictions, measure
required-variable coverage and guide/selector weights by time window, and
require exact original-tape parity. Full-input values and true variable masks
are diagnostic oracle information, never training inputs or a deployed rule.
Only after that evidence should a new interaction learner or joint-lookahead
observer be proposed and tested on fresh data. Do not lower the success
threshold or hard-code this parity target as a rescue.

## Verification and scope

- Generator truth tables, independent seed uniqueness and unchanged-bit v87
  integration parity: PASS under race instrumentation,3.005 seconds.
- Fresh generation: PASS,246.472 seconds. This is execution success, not
  adoption success; the evaluator separately returns FAIL.
- Evaluator verifies128 manifest entries,1,280 distinct base seeds, paired
  latent streams, configuration/order, feedback accounting and all metrics.
- `go vet`: PASS. Candidate code is unchanged; no new microbenchmark claim.
- Full deterministic replay: PASS,249.872 seconds. All2,560 records and
  prediction/latent hashes reproduce against the128-file manifest.
  `git diff --check` also passes.

The study does not cover every feedback schedule, outcome-dependent missingness,
arbitrary-dimensional events, real-agent outcomes or production concurrency.
All seven research directions remain open. No production, OpenClaw, commit or
push changes were made.
