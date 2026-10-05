# Fit-time age challenger v87

Frozen before fresh outcomes. Test a retained additional challenger, not removal
of the incumbent, count model or original 64-label subset model. v86 supports
old-window influence as a hypothesis, not an established causal rescue.

Use the same five v85 scenarios, four feedback schedules, 512-step streams and
64 trajectories per phase/cell. Fresh stream bases 2026118701/02; unchanged
frozen base, audit/latent/timing RNG roles. The 640 underlying trajectories are
paired across four schedules (2,560 runs), not independent repetitions.

Control is the retained-subset v85 arm. Candidate adds a subset model fitted
to the latest at most64 received live audits with origin > publication_step-128
and origin <= publication_step. Require16 eligible samples, otherwise fall back
to the original subset for that publication. The 128-step bound applies at fit
time; snapshots age between publications. No knowledge of true changes is used.
Minimum32 total received audits and cadence16 stay unchanged. One extra subset
fit is charged when eligible; no extra acquisition, labels or refit triggers.

Inner expert slots become [count, retained_subset, age_or_retained,
age_or_retained], with unchanged initial weights [.7,.1,.1,.1] and selector
updates. Duplicated slots aggregate weight; they are not independent evidence.
These are adaptive algorithm experts, not posterior probabilities that fixed
models are true. Fallback behavior belongs to the same declared algorithm.
All inner forecasts use the same actual observation mask. If the outer short
slot controls observation and non-count inner mass beats count, choose the age
observer only when its combined mass beats retained mass and it is available;
otherwise use retained. At most one observer runs, with the existing cost6 cap.
Model/version and split-generation invalidation remain unchanged. The control
with age disabled must reproduce the prior retained arm exactly.

Primary: candidate post-Brier gain >=.005 versus retained control with paired
z3.5 lower>0 for member/common under combined jitter/missingness, in both phases.
All other cells require full/post Brier harm upper<=.01 versus retained control.
Also report absolute Brier against constant0.5 (0.25), accuracy, feedback counts,
age support/fits and acquisition. Report all cells; any failed rule rejects the
pilot. A relative pass is not calibration or general delayed-learning success.

Require contract/race checks, no future/duplicate training origins, age-bound
and support tests, disabled parity, full deterministic replay, source hashes,
and isolated fit/foreground microbenchmarks. No production or serving change.
Further breadth, dependent missingness and real-agent validation remain required.
