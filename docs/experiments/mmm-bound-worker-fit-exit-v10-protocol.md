# Bound worker v10: fitted prediction after abrupt exit

Status: frozen before implementation, 2026-10-01. Research-only; production
unchanged. This extends the [v9 acknowledged-exit protocol](mmm-bound-worker-exit-v9-protocol.md).

Use the same persistent LibraVDB, durable-lineage wrapper, source log, and
motion worker as v9. Transfer one witnessed positive source label. Admit and
verify 31 additional negative worker outcomes at increasing observation and
feedback times, waiting until all 32 labels have been applied. This reaches
the learner's first component-fit boundary. At a later as-of time, journal an
unlabeled probe prediction and require its probability to differ from its
baseline by more than 0.01. Commit a future-only event, then exit the child
process without closing any component.

In the fresh process, reopen the same persistent state and motion worker.
Require 32 completed labels, one pending probe, an intact source index, and
next-ID continuation. Issue a new committed query at exactly the probe's
as-of time. Require matching features and baseline and **bit-identical**
probability to the pre-exit probe. If the inputs differ, the comparison is
invalid, not a pass. Missing or tampered source/lineage history must fail
closed under the existing v8/v9 guards.

Run focused race checks and the affected package suite. This finite test
establishes an acknowledged-write fitted-model restart only. It is not
hardware power-loss, uncertain-commit recovery, loaded p99, or evidence of
better agent answers.
