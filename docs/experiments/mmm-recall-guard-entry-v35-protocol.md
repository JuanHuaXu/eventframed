# Guard entry versus source validation v35: frozen protocol

Date: 2026-10-01. Research-only Goal 6 diagnostic after
[v34](mmm-recall-admission-guard-v34-results.md) found that
the pre-callback stage, not SQLite append or readback, dominates
serial admission at 4 ms offers.

Use the test-only `researchBoundSQLiteV26` guard wrapper and a
private context-tagged timestamp to divide v34's pre-callback
duration into:

1. admission start through entry into the guarded callback,
   including writer-gate wait, snapshot/lineage compatibility and
   wrapper overhead;
2. guarded callback entry through the durable-admission callback,
   i.e. `validateResearchCandidate` and its source reads.

Both nonnegative durations must sum exactly to v34's
`sourceValidation` duration for every selected label. No
production guard code, ordering, callback, durability, label
count or cadence may change. Run one fresh enabled-only 6 ms
fixture and one 4 ms fixture, each with the same 192 full
200-event Recalls, eight workers, 256 future-only writes,
64 bound labels, queue64, selected channel32, reorder cap32,
admission channel16, and guarded SQLite WAL/FULL journal.
Retain all zero-drop, exact nomination, no-future, as-of,
mutation rejection, journal reopen, durable replay and phase
conservation checks. Report actual offer gap, serving/freshness
p99, p50/p99 of both new stages, and stages for the five oldest
labels.

If guard entry dominates, pursue writer-gate contention or
admission scheduling; if candidate validation dominates, pursue
its source-read algorithm while preserving authority. Neither
outcome alone authorizes a behavior change. This diagnostic
does not establish a higher-rate pass or whole Goal 6 success.
