# Summary maintenance screen results

The predeclared screen FAILED: only one of four size/repeat comparisons met all
criteria. All initial/final graph hashes match between paired control/candidate
runs, and all four test commands passed. Verifier success checks artifact
consistency; its output `pass:false` is the screen outcome.

Run order was control0,candidate0,candidate1,control1 with CPU4. Each arm tests
N800 and6400,32 inserts and8 forced current-entry deletions after serial seeding.
Counters remain disabled in both overlays. Timers exclude vector generation,
state capture/hashing, and shutdown.

| Repeat | N | Initialization ratio | Insertion ratio | Deletion ratio | Screen |
| --- | ---: | ---: | ---: | ---: | --- |
| 0 | 800 | 1.0430 | 0.9877 | 0.9931 | fail |
| 0 | 6400 | 1.0100 | 0.9695 | 0.9089 | fail |
| 1 | 800 | 0.9855 | 1.0021 | 0.9784 | fail |
| 1 | 6400 | 0.9958 | 1.0043 | 0.7515 | pass |

Ratios are candidate/control total timed phase cost. Required: deletion<=0.90,
initialization/insertion<=1.05 in every comparison. The thresholds were recorded
before execution and are not relaxed to call the9.1% result a pass.

At N6400, eight deletion totals were1.107/1.385ms control versus1.006/1.041ms
candidate. These tiny totals and only two repeats do not support tail guarantees
or stable effect-size inference. Initialization and ordinary insertion did not
cross the screening regression limit, but that is not proof of zero overhead.

## Decision

Retain the summary as a correctness-tested architectural candidate, not a validated
throughput rescue. Removing registry scans is useful in principle as the registry
grows, but it is not established as the current end-to-end bottleneck. Do not spend
further tuning rounds merely to cross this small screen's threshold.

Return to private insertion/deletion preparation and bounded neighbor repair,
then measure integrated durability and sustained load. Larger-registry scaling
remains a future falsifiable test, not an inferred result. All seven whole goals
remain open. Raw outputs and the executable checker live in `summary-cost/` and
`check-summary-cost.mjs`; production and standard dependencies are unchanged.
