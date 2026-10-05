# Arrival-time routed learning v106: frozen protocol

Freeze before fresh outcomes. Retain v104's complete design: 12 cases, two
phases, 32 independent trajectories per phase/case, each run under immediate
and jitter0..31/missing0.2 feedback. There are768 underlying trajectories and
1,536 paired schedule-runs. No interim quality look, retuning, replacement seed
or optional enlargement. Use the same initial16 labels,256 scored steps,
step128 switches, model families, as-of fitting and32-step publication cadence.

Fresh seed base2110110600 + phase*1000000 + case*10000 + index*10. Roles0..4
remain rule/input/outcome/delay/missingness. Check effective seed separation
against archived v90-v104 learner/null allocations before generation. V105
reused v104 seeds and introduces no new allocation.

## One changed mechanism

Arms0..2 are unchanged generic64, conservative journal, and origin-prefix
role-carry journal. Arm3 is the arrival-time role selector: each delivered loss
updates selector weights once on arrival, while the evidence gate still drains
the version-scoped origin prefix. Prior, learning rate, per-label fixed share,
models, acquisition rule and deadlines are unchanged. No age discount or
event-clock forgetting is added to this experiment.

All arms share arrived-label training audits, never unrevealed outcomes. At t,
deliver prior due labels, expire unresolved age>=32 records, publish if due,
then acquire and forecast. Sample the current outcome afterward and deliver
zero-delay nonmissing feedback after forecasting. Flush clocks256..287. Keep
the last64/32 available labels ordered by origin, recording exact fit inputs.
Training audits are an additional nine coordinates per arrived label plus16
initial labels, separate from foreground cost.

## Frozen quality gates

Use expected Brier with paired mean +/-3.5 standard errors over32 trajectories
within phase/case/schedule/segment. Smaller is better. These are approximate
fixed-sample screens, not confidence sequences. Full256 and late128 segments
are kept. Schedules sharing a trajectory are not independent replicates.

- 288 non-harm gates: arrival versus each of the three controls, every phase,
  case, schedule and segment; upper Brier harm<=.01.
- 20 gain gates versus generic: all-step parity3/4/complement4 and late two
  switch cases, both phases and schedules; mean gain>=.005 and lower>0.
- 6 gain gates versus conservative: delayed late parity4 and both switch cases,
  both phases; mean gain>=.005 and lower>0.
- 4 additional gains versus full carry: delayed late two switch cases, both
  phases; mean gain>=.005 and lower>0.

Overall PASS requires all318 gates and integrity checks. V104's218 questions
remain represented for the new candidate, with additional full-carry protection
and recovery comparisons. V104 itself stays failed regardless of this result.
Report expected accuracy, realized Brier, paid coordinates and update counts.
Do not claim zero harm from a passed .01 allowance.

## Integrity

Before fresh generation, reproduce all original v104 control outputs on
consumed design cases0,3,10,11/index0 under both schedules. Immediate arrival
and full carry must match exactly. Preserve step-level data, per-arm journal
settlement and selector update counts. Arrival selector updates can exceed
settled-prefix counts before flush; independently reconstruct both clocks.

Require paired latent tapes, exact as-of training origins, no duplicate updates,
no stale-version gate updates, bounded foreground/journal cost, deterministic
replay of every schedule-run, source hashes, and independent metric/accounting
reconstruction. Benchmark only after quality/replay processes complete.

This is not a delayed test-martingale theorem or a production evaluation.
Informative censoring, abstraction-generation changes, broader v88/v80 outcomes,
independent generators, real agent tasks and persistence-tail requirements
remain separate. No production changes, commits or pushes are authorized.
