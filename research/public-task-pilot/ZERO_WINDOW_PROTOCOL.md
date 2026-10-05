# Zero deliberate coalescing delay: synchronous-WAL ablation

Freeze before dispatch. Copy pinned libravdb1.6.13 into an isolated research
directory; use research-zero-window.mod/sum, leaving normal go.mod/sum and
installed modules unchanged. Go disallows overlays below GOMODCACHE, so the
initial shared-cache overlay is unused. The local-source overlay changes only
groupCommitWindow initialization from1ms to0. Adaptive delay returns0 at that
setting. Do not change file.Sync, flushBatch, durability mode, acknowledgement
ordering, index mode, service gates, deadlines or retry counts.

Run the dependency's synchronous-default, group-sync acknowledgement,
sync-failure propagation and batch-recovery tests with the overlay first.
These finite tests complement source inspection; they are not arbitrary crash
or hardware durability proofs. Some tests override grouping windows themselves.

Then repeat the full32-arm DURABLE_OPENLOOP_PROTOCOL grid without instrumentation:
200 events,32reads+16writes, lower/higher fixed-arrival rates, admission off/on,
ordinary/experimental packing, future/visible writes, two repetitions. Use the
incremental diversity service variant and local storage; fresh retained stores.
All prior correctness,100ms, warmup and orderly-reopen gates remain unchanged.

Compare with the earlier durable artifact descriptively, since executions are
not simultaneous. If any admitted arm fails, the full-grid rescue fails. A
lower deliberate wait may increase flush count and contention; do not assume
zero is optimal. This experiment does not measure power or flush counts.
No candidate parameter tuning on its outcomes. Any winning setting still needs
longer/randomized-arrival and crash/cancellation testing. No production changes.
