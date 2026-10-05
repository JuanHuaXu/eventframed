# Matched offered-load diagnostic v57

Frozen before measurement. v56 found an isolated admission/retry cost reduction,
but v55 remains a failed loaded rescue. Its producers are closed loop: four
readers send the next request after completion and the writer sleeps2ms after
each completed write. Different arm costs therefore change offered arrivals.
This diagnostic measures a separately declared fixed offered schedule; it does
not repair or replace v55 and does not modify any runtime implementation.

Use the same four arms: off; raw prepared durable; batch source with combined
verified cleanup; resolved source with combined cleanup. Rotate four arms in
each of three trials for each read interval2/5/10ms, with corresponding write
interval4/10/20ms:36 cells. Each cell offers192 recalls and96 future writes.
Read index i*4+w stays in lane w; all due times share one monotonic origin.
Retain four bounded reader goroutines and one writer. A late call keeps its
original due time, including backlog behind a preceding call in the same lane.
No unbounded goroutine generation, schedule resetting, or dropped offered calls.

Record index,due,start,end for every read/write, with unique complete index sets
and inside-call duration=end-start. Due-to-end includes producer/lane lateness;
start-minus-due reports that lateness separately. The origin is published by
the start-channel barrier. Context cancellation must terminate waits and join
all producers/consumer. All ordinary call/queue/guard/original/terminal counts
remain mandatory. Old constructors and nil-schedule fixtures remain controls.

Keep public50-event overlap,K50/pack10,fixed as-of,queue64,group<=4,20ms guard
entry budget,FULL durability,parent callback context. No labels/fitting/private
data/production. All workload cells drain completely before the next cell.
Capture raw timings and source/hash snapshots in an exclusive artifact. Run
timing-invariant/cancellation and actual fixture accounting/race checks first.

Report all rates/trials and all arms. For each resolved cell: require age p95
<=250ms,192 accepted observations, and both read and write due-to-end p99
<=1.10x same-rate same-trial off. This explicit offered-load non-harm screen is
not the older read-only screen. Report inside-call p99 alongside scheduled
latency and lateness; never add unpaired quantiles. Accepted-observation age
still begins at frontier enqueue, not at offered request time. No total
offered-to-background-completion metric is claimed without per-request linkage.

Zero unexpected errors or original mismatches. Go PASS means accounting, not
automatic screen success. All trials retained even if off itself accumulates
backlog. Passing only a slower rate is a finite capacity finding, not a rescue
at higher rates or completion of the seven-direction goal. No threshold tuning,
default changes, whitepaper accuracy changes, push or deployment.
