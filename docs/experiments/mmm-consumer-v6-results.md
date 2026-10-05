# General adapter consumer timing v6

All18 diagnostic arms finished; the frozen screen FAILED four cells. Raw source
and measurements: mmm-consumer-v6.jsonl. The test exited1. Independent audit
verified source hashes/current source, unique arms, read/write counts, stage
sample counts, admission/drop conservation and50 completed labels per frontier.
All arms had no errors, no worker failures and overlapping reads/writes.

For192-request arms, mean milliseconds per admitted frontier:

| Queue / trial | Admission | Feedback submission | Completion wait | Poll count |
| --- | --- | --- | --- | --- |
| 16 / 0 | 1.9576 | .0412 | 3.9743 | 2.981 |
| 16 / 1 | 1.9440 | .0399 | 3.9712 | 2.974 |
| 16 / 2 | 2.2176 | .0460 | 3.7860 | 2.922 |
| 64 / 0 | 2.0775 | .0689 | 3.8787 | 2.979 |
| 64 / 1 | 1.9529 | .0402 | 3.7675 | 2.911 |
| 64 / 2 | 1.7922 | .0407 | 3.7679 | 2.865 |

Completion wait is the largest measured consumer stage. It includes fitting and
other useful worker computation, scheduling and polling overshoot; these data
do not separate them. Approximately three1ms polls occur per frontier. This is
not evidence that3ms is avoidable, and it is not a CPU profile.

Failures: short queue16 trial1 serving-p99 ratio40.013/33.154 exceeds1.10;
long queue16 trial2 admits153/192 below154 required; long queue64 trials0/1 have
age p95 270.601/262.864ms above250ms. Other cells passed the frozen criteria.
Instrumentation can perturb scheduling; retain all prior failures and avoid
inferring a new regression from separate runs.

Next lead: a bounded completion-notification interface, retaining the completion
counter and cancellation/close semantics, compared against the unchanged polling
harness. It must avoid lost wakeups and must not publish future/outcome data into
the original journal. Unit/race checks must precede another load experiment.
This is a hypothesis, not an implemented or validated rescue. Accuracy and other
research directions remain open; production is unchanged.
