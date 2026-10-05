# Delayed routed learning v104: frozen protocol

Freeze before fresh quality outcomes. No interim looks, tuning, seed replacement
or optional enlargement. Two phases, 32 independent trajectories per phase and
each of the 12 v103 cases. Each of 768 underlying trajectories runs under both
immediate feedback and jitter0..31/missing0.2: 1,536 schedule-runs, not 1,536
independent latent trajectories. All generators, rule pools, 16 initial labels,
256 scored steps and step128 changes remain as in v103.

Seed base2100110400 + phase*1000000 + case*10000 + index*10. Roles0..4:
rule, input, outcome, delay, missingness. Delay and missingness are independent
of outcome and consume the same RNG draws in both schedules. Audit effective
seed separation against the v90-v103 learner/null allocations.

## Arms and chronology

0: generic64 with archived one-step observation.
1: coherent routed one-step with conservative version-scoped journal.
2: coherent routed one-step with stable-role selector carry.

All arms receive the same arrived full-frame training audits and use the same
four fitted models: generic64, Boolean64, generic32, Boolean32. Fit every32
origins, selecting the last64/32 labels by origin among arrived evidence, plus
eligible initial labels. Record every fit's exact input origins. Uniform input
law stays explicit; dependent4 intentionally misspecifies it.

At clock t: deliver previously issued labels due by t, expire unresolved origins
with age>=32 (already delivered buffered labels survive), then publish if due,
then acquire/forecast the current frame. Only afterward draw its outcome and
deliver it immediately if delay0 and not missing. This order, including expiry
before publication, is fixed. Flush clocks256..287 without further forecasts or
publications. A label's availability for fitting is its arrival, not when a
blocked selector journal eventually drains it. Training audits are separately
counted as nine coordinates per arrived label plus16 initial labels.

The full six-coordinate observation cap, forced first view, stopping rules,
bank, routing and falsification kernels are unchanged. A stale role loss may
update the selector but cannot update the current model's evidence tests. Roles
must retain their meanings. No abstraction split is introduced in this run.

## Frozen quality gates

Paired per-trajectory intervals within phase/case/schedule/segment are mean
+/-3.5 sample standard errors over32 trajectories. These are approximate fixed
sample screens, not confidence sequences or population guarantees.

- 192 non-harm gates: role carry versus generic and versus conservative, all12
  cases, both phases, both schedules, full256 and late128; upper Brier harm<=.01.
- 20 gain gates versus generic: all-step parity3/4/complement4 and late two
  switch cases, both phases and schedules; mean gain>=.005 and lower>0.
- 6 incremental carry gates versus conservative: delayed late parity4 and both
  switch cases, both phases; mean gain>=.005 and lower>0.

Overall PASS requires all218 gates and integrity checks. Do not weaken them
after results. Expected Brier under simulator truth is primary; report expected
accuracy, realized Brier, acquisition cost and feedback classifications too.
Carrying more losses is not itself a quality pass, and immediate and delayed
versions of one trajectory must not be pooled as independent samples.

## Integrity

Before fresh generation, immediate arms must exactly reproduce v103 generic
and routed one-step metrics on consumed design cases0,3,10,11/index0. The two
journal variants must be identical there. Preserve v103 artifacts.

Store full scored steps, delay/missingness, per-step journal accounting, final
counts and exact fit origins. Reconstruct scores, training eligibility and
origin-prefix settlement counts independently. Replay every run and verify
paired latent tapes. Hash protocol, evaluator, driver and model dependencies.
Require no pending records after flush, bounded capacity, original forecast
binding and no current-version gate updates from stale records.

The independently randomized schedules do not establish a general delayed
likelihood-ratio certificate. Informative delays, arbitrary dependent outcomes,
member-split generation changes, independent generators, production serving and
the older broader v88/v80 criteria remain outside this first quality comparison.
