# Deadline trace: most work enters nearly expired

Six actual service arms, three paired direct/batch trials. Existing100ms age,
capacity16, temporal compatibility and all-distinct workload unchanged.

| Arm | Trial | Processors started | Entry <25ms | Interrupted | Completed | Successful fit ms |
|---|---:|---:|---:|---:|---:|---:|
|Direct|0|23|21|22|1|71.234|
|Batch|0|20|17|19|1|58.444|
|Direct|1|21|18|20|1|80.209|
|Batch|1|23|22|22|1|60.811|
|Direct|2|20|17|19|0|79.158|
|Batch|2|25|24|24|1|59.768|

Batch totals:63 of68 entered with less than25ms remaining;65 had less than40ms;
65 were context-interrupted. This supports late admission to computation as an
important failure mechanism. Successful batch fits take58-61ms under this load,
not the earlier isolated43ms. These successful-fit timings are selection-biased,
not certified lower/upper bounds or an average over canceled full fits.

Every completed processor result matched its reference. One direct result was
rejected after returning successfully; this trace cannot distinguish store
compatibility change from elapsed deadline in that post-check. Accepted jobs
that never entered processing number18-25 per arm. Drops remain19-24 per arm.
Zero read/write errors and16 overlapping writes per arm. All64 offered jobs
are accounted for, including work that never produced a result.

Two entry budgets were slightly negative despite the scheduler's earlier
check. Scheduling between a check and execution can cross a deadline; the
fitter's own context checks still rejected these jobs. This is not evidence
that the scheduler published expired results.

## Boundary and next action

The100ms minus entry-remaining budget combines queue waiting, compatibility
checks and scheduling. It is not an isolated queue-wait measurement. No new
timestamp instrumentation was added inside production queue/store code.
Trace recording occurs before post-fit validation and may influence tight jobs.

The next controlled candidate can decline a fit whose remaining deadline is
below a predeclared computation allowance. Such rejections must remain in the
offered-work denominator and cannot be called completions. Compare actual
completion and foreground effects, not just fewer interrupted fits. This may
reduce wasted work without meeting52/64; do not promise a capacity rescue or
relax the original screen. An admission rule is not a substitute for improving
the actual learner's usefulness or forecasting quality.

Raw traces and source hashes: `segment-deadline-trace-results.json`. Verify
via `node research/verify-segment-deadline-trace.mjs`; it reconciles IDs, trace
counts, timing identities and terminal accounting. It does not certify the
unobserved post-check cause. Production remains unchanged; all directions open.
