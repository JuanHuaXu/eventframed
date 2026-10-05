# Bound worker v9: acknowledged abrupt-exit recovery

Status: frozen before implementation, 2026-10-01. Research-only; production
unchanged. This extends the [v8 motion contract](mmm-bound-worker-motion-v8-contract.md).

Use a child process with persistent LibraVDB, durable publication lineage, a
durable source log, and a motion-mode bound-worker log. In the child:

1. Commit an event and an observed query. Admit one source-bound forecast with
   a witness, commit verified feedback, and wait until the label is applied.
2. Prepare and open the motion worker. Admit one more witnessed forecast from
   a committed query, commit verified feedback, and wait until it is applied.
3. Commit an event available only after the bootstrap cutoff, then exit the
   process without calling Close on the service, stores, or logs.

Only the expected child exit code counts as reaching the boundary. In a fresh
process, open the same backend and lineage in recovery mode, reopen both logs,
rebuild and reopen the motion worker against the recovered target. Require two
completed labels, source-index recovery for the worker's original, and next-ID
continuation. A pre-cutoff backfill after recovery must reject another motion
handoff. Any missing as-of proof, source witness, origin seal, or journal state
must fail closed, not silently start a fresh epoch.

Run focused race checks and the affected package suite. This is an
acknowledged-write abrupt-process-exit test, **not** hardware power-loss,
uncertain-commit, cross-database atomicity, loaded p99, or agent-answer proof.
