# Delayed/missing retained learning v85

Frozen before fresh streams. Compare the unchanged count and retained-subset
models through the v84 journal under four schedules: immediate; fixed16-step
delay; independent20% missing with zero delay; uniform0..31-step delay plus
independent20% missing. Schedule draws use a separate RNG role and do not depend
on features, outcomes, forecast confidence or audit selection.

Use the five v82 scenarios and frozen4096-label base. Fresh stream bases
2026118501/02, existing Seed role separation,64 streams per scenario/phase.
The same640 underlying trajectories are reused across the four schedules for
paired comparisons:2560 schedule-runs, not2560 independent underlying streams.
Each run has512 forecasts per arm. Shared latent/audit streams do not change
when feedback timing changes; this is tested against immediate v82 composition.

Each step expires origins older than48, issues both forecasts and acquires any
predeclared25% audit views, then creates simulator labels and scores emitted
forecasts. Labels enter the learner only when their packet is due. Due packets
are processed in origin order within each arrival batch. There is no extra
terminal training period; remaining journals are explicitly censored at512.
Never use missing or not-yet-arrived labels for monitoring, fitting or weights.

The frozen base's paired correctness forecasts are stored with each packet.
The Bayesian Monitor and paired investigator receive only delivered labels,
in arrival order, once each. Nomination plus investigator evidence authorizes
the research split proxy. Journal stale/version checks still apply. Both arms
must have identical applied/stale/censored counts and split timing.

Received audit pairs are sorted by event origin; retain at most256, fit count
short and subset on latest64, local on retained live audits, pooled on latest128
live/reference audits. Minimum32 received audits and cadence16 remain unchanged.
Publish immediately after the triggering packet's fit, before later packets
in that arrival batch. Old-version selector feedback is skipped, but a received
label may still enter training once. Record these two uses separately.

Each member/common-shift case and schedule must have candidate post-Brier gain
>=.005 versus its equally delayed count control, with paired z3.5 lower>0.
Other scenarios require full/post harm upper<=.01. Report both phases and every
cell, including absolute degradation versus immediate. Any failed cell rejects
overall adoption. These are conditional finite trajectory screens, not arbitrary
missingness, calibration, false-split or production guarantees.

Require immediate-composition parity, deterministic replay, no label invention,
pending/accounting bounds, actual cost caps, and archive source hashes. Measure
statistical quality here; v84's lifecycle tests alone are not learning evidence.
