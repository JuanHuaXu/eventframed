# Learned contrast v3 retrospective model/schedule factorial

This is a **post-hoc diagnostic on consumed v3 tapes**, not a new confirmation
test or a rescue pass. Inputs are the immutable v3 design and confirmation
archives, SHA256 `0dcd2eae7c3ed64ee43844e731fba7d07c850f77cdfd7a165e59f4e30ba5fb59`
and `856b8fb095aca8b374be471ea6034d630640b3cf12e280c6837c38a839efc2ac`.
Keep both splits and every scenario; do not tune a parameter on either.

## Four replay cells

For each tape, take the actual selected-clock schedules from the 11-rule
learned arm and the 12-rule learned arm. Replay each schedule through both
forecasters with its originally declared prior, rolling-32 likelihood and
origin-feature/delivery semantics:

| | Old learned schedule | Null-augmented learned schedule |
| --- | --- | --- |
| Old 11-rule forecaster | `old/old` | `old/new` |
| New 12-rule forecaster | `new/old` | `new/new` |

All cells forecast before any current-clock delivery; only a selected,
nonmissing, due label enters the corresponding model. `old/old` must reproduce
archived arm 2 and `new/new` archived arm 5 at every one of 512 clocks. Exact
requested-label counts, due times and archived scores must also agree. Abort
if either diagonal fails. Crossed schedules are externally supplied in replay
and need not be causally implementable by that forecaster; never describe them
as live policies.

## Diagnostic contrasts

Report paired means and independent-fit-cluster mean +/- 3.5 SE for full,
post and evaluator-only expected post Brier, plus the v3 restricted recovery
delay and miss rate. On the same old schedule, `new/old - old/old` isolates
the model-family effect. Within the old forecaster, `old/new - old/old`
isolates the acquired-label schedule effect. Compare `new/new - new/old`
for the schedule effect under the new forecaster; their difference is a
model-by-schedule interaction. Report immediate and delayed bit2, bit0,
interaction, null and majority without suppressing counterexamples.

This diagnosis can rank mechanisms for a **newly frozen** rescue study. It
cannot revalidate v3, establish selection causal effects in real agents,
give Anti-Pigeon authority or justify tuning on these consumed cohorts.
