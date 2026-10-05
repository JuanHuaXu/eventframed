# Bound worker v9: acknowledged abrupt-exit recovery

Date: 2026-10-01. Frozen [v9 protocol](mmm-bound-worker-exit-v9-protocol.md).
Research-only; production unchanged.

The subprocess screen **passes** on persistent LibraVDB plus durable
publication lineage. The child acknowledged one source-bound verified label,
one new guarded worker label, and a future-only event, then exited without
closing the service or logs. A fresh process reopened all three persisted
stores, rebuilt the same motion epoch, recovered two completed labels and the
source index, and continued with the next prediction ID. A later pre-cutoff
backfill rejected another handoff. The child had to exit with code 27; an
ordinary test failure did not count as reaching the crash boundary.

Reproduce with:

```sh
go test ./internal/service -run '^TestResearchMotionBoundWorkerAcknowledgedAbruptExit$' -count=1 -v
```

This is **acknowledged-write abrupt-process-exit** evidence, not hardware
power-loss or uncertain-commit recovery. The two-label fixture is below the
learner's 32-sample component-fit threshold, so this result does not prove a
fitted model gives the same prediction after restart. It also does not measure
loaded p99, cross-database atomicity, visible-mutation state transfer, or
untouched agent outcomes. Goal 6 remains OPEN.
