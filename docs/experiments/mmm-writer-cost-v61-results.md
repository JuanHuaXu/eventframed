# Bounded native writer concurrency v61 results

## Verdict

The frozen isolated screen FAILS. Four submitters improve whole-workload time
by 44.76-46.45% in all three trials, but sequential p99 regresses 11.95% in
trial 0, exceeding the 10% ceiling. Do not tune that ceiling after measurement.
This is a promising concurrency lead, not a validated serving rescue or a
completed research direction. Default Open still uses two slots.

## Evidence

[Protocol](mmm-writer-cost-v61-protocol.md),
[raw artifact with source snapshots](mmm-writer-cost-v61.jsonl).

Artifact SHA256:
`03397cf76e898d3ae8a2aec6682908b1e79fabdced5358d16fe8b204f0b3c6c9`.
All eight captured source hashes were recomputed against both embedded sources
and local files. Go1.27.1, darwin/arm64, Apple M4, GOMAXPROCS10. Twelve cells,
128 journals per cell, 1536 journals reopened and fully compared. Each cell
stores 4416200 JSON payload bytes; paired complete-record digests and payload
sizes match. Fifty placeholder decisions per journal are storage-shaped public
data, not predictions or learning evidence.

| Submitters | Trial | Two-slot elapsed ms | Four-slot elapsed ms | Elapsed improvement | Two-slot p99 ms | Four-slot p99 ms |
| --- | --- | --- | --- | --- | --- | --- |
| 1 | 0 | 714.304 | 687.853 | 3.70% | 6.238 | 6.984 |
| 1 | 1 | 688.520 | 715.765 | -3.96% | 6.778 | 6.349 |
| 1 | 2 | 718.572 | 711.855 | 0.93% | 6.812 | 6.929 |
| 4 | 0 | 336.413 | 180.144 | 46.45% | 17.998 | 6.544 |
| 4 | 1 | 342.006 | 188.919 | 44.76% | 12.077 | 6.833 |
| 4 | 2 | 334.225 | 182.735 | 45.33% | 12.196 | 6.177 |

Per-call p99 uses nearest rank ceil(0.99*n), n=128. Whole-workload timing runs
from shared start release through all submitting goroutines joining. Timing
excludes payload construction, reopening and verification; calls include
synchronous durable acknowledgment. Trials are diagnostics, not simultaneous
confidence bounds or independent deployment confirmations.

## Safety and scope

- Native-store race suite PASS (6.247s); native-store vet PASS.
- Experiment PASS (7.569s package time) means integrity, not the performance
  screen. Every cell also accepts exact retry and rejects conflicting contents.
- Separate abrupt-exit test verifies 16 acknowledged concurrent journals after
  child process exit without Close/checkpoint. This is not a power-loss test.
- Private constructor helper preserves ordinary Open's two-slot limit; the
  research constructor alone selects four. Queue depth64, snapshot/identity
  locks, synchronous WAL durability and synchronous indexing are unchanged.
- Pinned dependency has a 1ms initial coalescing window, 500us step and 5ms
  maximum, with immediate flush requests and a 10ms fallback. These are internal
  defaults, not a public tuning API. No dependency/module-cache edits were made.

More overlapping durable requests may improve grouping, but this experiment
does not identify the exact internal mechanism or prove mixed event/journal
writer non-harm. Next compare the opt-in constructor under the same scheduled
service arrivals, including plain native controls, active shadow work, writer
backlog, observation completion and snapshot correctness. Retain the sequential
failure; do not promote the constructor or claim that v57-v60 are rescued.
