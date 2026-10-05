# Sparse pilot checkpoint: allowance stop

The account tool reports76% of the weekly allowance used, exceeding the user's
explicit75% stopping threshold. Research stopped before starting collection.
All seven research goals remain OPEN; no quality result is implied.

Completed this turn:
- Frozen504-fit pilot protocol (index0, all phases/scenarios/schedules,
  clocks0/128/224,64/32-label windows).
- Implemented `runSparsePilot` and opt-in `TestSparsePilotCollect` in
  `internal/observationlearners/sparse_vrvm_pilot_test.go`.
- Added `gaussianXi` snapshot metadata to the fitter: the exact xi used for
  the final Gaussian, before its subsequent coordinate update.
- `TestSparsePilotAsOf` passes under race (4.262s package time): unavailable,
  future and evaluator-data changes leave forecasts unchanged; admitted-label
  changes affect forecasts; complemented labels give complementary laws;
  the recorded Gaussian input state reconstructs the fitted mean.

NOT done:504-fit pilot collection, independent pilot audit, pilot replay or
quality scoring. No collector process is running. Earlier sparse numerical
results are not evidence of trajectory accuracy. The Gaussian snapshot field
is an audit-metadata addition, not a fit-equation change.

Resume only when the allowance boundary permits or the user changes it. From
the repository root, the prepared collection command is:

```sh
EVENTFRAME_SPARSE_SOURCE=../../docs/experiments/mmm-soft-learners-v120.jsonl EVENTFRAME_SPARSE_OUTPUT=../../docs/experiments/mmm-sparse-vrvm-v1-pilot.jsonl go test ./internal/observationlearners -run '^TestSparsePilotCollect$' -count=1 -timeout=15m -v
```

First verify the output path is unused and no prior collector is live. Preserve
all old controls and failures. Independent audit should rebuild Gaussian state
from UsedXi and PriorVariance, not FinalXi; include numerical integration and
all recorded fit-origin checks. One trajectory per cell does not support the
full32-trajectory quality intervals. No production/paper/install/commit/push
changes were made.
