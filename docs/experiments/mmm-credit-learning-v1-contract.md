# Credit-controlled monitoring: closed-loop forecast test

Use the unchanged sequential inner-arrival learner with the frozen credit-ledger
monitoring policy. Keep all four internal learning arms, but the primary comparison
is full arrival-learning (arm2) with credit monitoring against the SAME arm2 from
the original switch-transfer artifact. No additional pooling, fitting, learning
rate, residual, threshold, or forgetting changes are permitted.

Replay128 consumed trajectories, immediate and jitter0..31/missing.2 schedules.
All latent tapes, audit nominations, fit-origin manifests and label availability
must match the original. The monitor's split clocks and512 credit transitions
must reproduce the shadow-credit artifact exactly. The disabled policy must
match the old driver on representative cases. Live forecasts are emitted before
the same-tick outcome; neither parent nor shadow predictions supply runtime input.

Before collection freeze: on the four reverse-direction cells (two cohorts times
two schedules) require mean post Brier gain>=.005 and mean-3.5SE>0. Require upper
full/post harm<=.01 on ALL16 case/cohort/schedule cells. These consumed-data
fixed-sample screens do not provide research-wide coverage. Report relative and
absolute Brier, correctness, foreground cost, monitor/audit cost, split clocks,
and exact ledger prefixes. New foreground choices may change total acquisition
cost even though monitoring alone has a prefix guarantee; do not hide this.

No fresh confirmation, real-agent utility, universal conditional-law certificate,
serving-latency validation, or whole-goal completion follows from this run.
