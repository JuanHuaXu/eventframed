# Kalman early-recovery phase profile (post-hoc)

This diagnostic partitions the already-consumed [v2 Kalman study](mmm-kalman-early-v2-results.md)
into four 16-clock segments after each change. It is **not** an independent
confirmation, a tuned policy, or a Goal 2/4 pass. The [profile script](../../research/kalman-phase-profile-v1.mjs)
pins the two source hashes, checks cohort sizes, and recomputes mean Brier from
every issued forecast and realized outcome. The existing
[v2 verifier](../../research/kalman-early-v2-verify.mjs) separately checks
forecast/feedback ordering and the frozen v2 decision.

Each cell is confirmation mean Brier for `last64 / q=.0005 / q=.01` over 32
trajectories. Lower is better. The design cohort shows the same pattern.

| Clocks after change | Immediate shift | Shift with delay 16 and 25% missing |
| --- | --- | --- |
| 0-15 | .24913 / .37717 / .36326 | .25095 / .42036 / .41666 |
| 16-31 | .25217 / .32955 / .26099 | .25073 / .41574 / .41636 |
| 32-47 | .25160 / .25978 / .18202 | .25146 / .36170 / .31222 |
| 48-63 | .25363 / .23641 / .15111 | .25296 / .28271 / .22773 |

In the delayed scenario, **zero** post-change audited labels arrive during
clocks 0-15 in either cohort. The confirmation cohort receives 77 such labels
across 32 trajectories during clocks 16-31, then 100 and 107 in the next two
segments; design counts are 83, 89, and 80. Because the harness forecasts
before processing deliveries at a clock, a label arriving at clock 16 can
first affect the forecast at clock 17. In the immediate-shift confirmation,
134 post-change audited labels arrive during clocks 0-15.

This supports a narrow diagnosis: more process noise improves later adaptation
after evidence arrives but cannot identify the new rule during a period with
no relevant labels. The working filter also stays too confident relative to
the near-.5 last-64 control in that period. It does **not** prove that all
state-space models fail; a predeclared external onset cue, a credible change
hazard, or uncertainty-aware fallback could change the information available
to a new candidate. Those are new hypotheses requiring fresh, frozen cohorts,
stationary protection, equal observed-label budgets, and full cost accounting.
Do not reinterpret this post-hoc slicing as independent evidence for a rescue.

Reproduce with `node research/kalman-phase-profile-v1.mjs` and
`node research/kalman-early-v2-verify.mjs` from the repository root. The latter
prints the full v2 audit and must end with `"overallPass": false`.
