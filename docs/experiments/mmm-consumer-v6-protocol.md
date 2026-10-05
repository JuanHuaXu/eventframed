# General adapter consumer timing v6

Diagnostic fork of frozen v4, retaining all18 arms, queue sizes, workloads and
pass criteria. Add per-frontier admission, feedback-submission and remaining
completion-wait durations, plus number of1ms completion polls. Do not alter the
worker, polling schedule, thresholds or production configuration.

Waiting includes useful concurrent worker computation and scheduling delay;
poll count times1ms is not wasted CPU or proven avoidable time. Stage timing is
descriptive and may perturb scheduling. It cannot by itself separate fitting
from idle overshoot. Any follow-up notification rescue needs its own comparison.
Raw source/hash snapshots and arm measurements are exclusive-created as JSONL.
