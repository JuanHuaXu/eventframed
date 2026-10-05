# Scheduled native writer concurrency v62 results

## Verdict

FAIL overall: four-slot resolved source passes the frozen joint screen in only
one of three trials; two-slot resolved source passes none. Four slots improve
active scheduled writer p99 and observation age in all three paired trials, but
writer non-harm against matching native controls still fails trials0 and1.
No default promotion, completed direction or general latency claim follows.

## Reproducible evidence

[Protocol](mmm-writer-load-v62-protocol.md),
[measured artifact](mmm-writer-load-v62-run2.jsonl).
SHA256 `58ada118178ee9c492387e8a481b240adad46de68306a58983c8de625c739800`.
All364 captured source hashes match both embedded content and local files.
Go1.27.1, Apple M4, darwin/arm64, GOMAXPROCS10; no concurrent task-started
tests/profiling during measurement. No production or dependency changes.

The first invocation failed before source capture or any measured cell because
filepath.Rel rejected unresolved parent-relative paths. The original
mmm-writer-load-v62.jsonl is retained as a zero-byte preflight failure. Capture
now resolves an absolute repository root; no workload or threshold changed.
The second invocation is the sole measured run, not a selected successful rerun.

Full service race suite PASS67.143s before the source-capture repair. Focused
four-arm accounting race test PASS5.019s and vet PASS after that repair. The
measured experiment PASS19.522s means integrity/accounting, not the joint screen.

## Results

All times are ms. Read/write p99 include original offered-time lateness.
Age p95 is from frontier enqueue, not original request due time. Nearest-rank
percentiles; each cell has192 reads and96 writes at5ms/10ms offered spacing.

| Trial | Arm | Read p99 | Write p99 | Write lateness p99 | Age p95 | Joint screen |
| --- | --- | --- | --- | --- | --- | --- |
| 0 | Native two | 568.765 | 107.385 | 94.023 | n/a | control |
| 0 | Native four | 603.593 | 153.799 | 141.815 | n/a | control |
| 0 | Active two | 38.348 | 350.469 | 340.469 | 131.662 | FAIL |
| 0 | Active four | 26.130 | 231.437 | 221.437 | 85.792 | FAIL |
| 1 | Native two | 578.628 | 112.900 | 98.274 | n/a | control |
| 1 | Native four | 425.621 | 68.921 | 58.921 | n/a | control |
| 1 | Active two | 35.438 | 337.413 | 327.413 | 119.151 | FAIL |
| 1 | Active four | 28.282 | 185.451 | 175.451 | 87.853 | FAIL |
| 2 | Native two | 513.657 | 60.818 | 50.819 | n/a | control |
| 2 | Native four | 611.643 | 139.955 | 125.898 | n/a | control |
| 2 | Active two | 33.536 | 339.717 | 329.717 | 112.596 | FAIL |
| 2 | Active four | 29.824 | 120.943 | 110.943 | 88.336 | PASS |

Four-slot active/native writer ratios are1.505,2.691,0.864, against ceiling1.10.
All active cells meet the read ratio,250ms age and completion requirements.
Native-four versus native-two varies in both directions; do not infer that the
constructor alone solves native-serving backlog. Active-four inside-call writer
p99 is29.917/26.804/24.058ms; the much larger scheduled values expose accumulated
producer lateness. A short individual call is not a recovered offered-load tail.

Across12 cells:2304 reads,1152 event writes. Each active cell accepts192 frontier
observations. All57600 source originals are read back and matched, with57600
terminal discards, no drops, expiry or recorded errors. Checks retain snapshot,
identity, group-size and phase-containment invariants. Observations are cold and
untrained; this measures neither accuracy nor loaded training cost.

## Interpretation and next lead

The v61 concurrent-write lead carries into mixed background load, but an isolated
45% speedup was not sufficient for reliable joint non-harm. Remaining write
backlog and variable native controls require attribution, not another threshold
change. Next isolate four-slot journal contention against event ingestion and
publication-gate waiting with arm-specific measurements. Preserve durability and
snapshot compatibility through commit; do not release native journal locks early
without an equivalent correctness contract. Default Open remains two slots.
