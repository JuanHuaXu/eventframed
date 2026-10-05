# Buffer rescue v2: failed longer-load screen

**Queue64 does not pass the frozen advancement rule.** It absorbs the64-request
burst, but the192-request workload still loses work and exceeds the learning-age
budget. Queue16 also fails throughout. No production queue setting was changed.

See [protocol](mmm-learning-buffer-v2-protocol.md) and
[18-arm raw artifact](mmm-learning-buffer-v2.jsonl). Three trials per workload,
off/16/64 rotated order, fresh temporary persistent stores, concurrent readers,
future-only writers, and actual50-label model updates per admitted frontier.

| Requests | Trial | Queue64 completed | Completion p95 ms | Queue16 completed | Completion p95 ms |
| --- | ---: | ---: | ---: | ---: | ---: |
| 64 | 0 | 64/64 | 244.701 | 32/64 | 223.126 |
| 64 | 1 | 64/64 | 249.701 | 33/64 | 215.568 |
| 64 | 2 | 64/64 | 248.587 | 33/64 | 219.855 |
| 192 | 0 | 115/192 | 790.304 | 67/192 | 293.875 |
| 192 | 1 | 116/192 | 793.693 | 68/192 | 281.710 |
| 192 | 2 | 118/192 | 792.595 | 69/192 | 283.612 |

Every admitted label completed; no read/write/bridge errors were reported.
Larger buffering raises longer-load completion to59.90%-61.46%, still below80%,
and pushes completion age far beyond250ms. Short-burst completion ages are just
under that ceiling, not a robust margin. Short-burst trial1 queue64 additionally
fails the serving-p99 allowance (33.043ms versus29.707ms off).

Queue64 can retain at most3200 candidate projections for these50-candidate
frontiers, versus800 at queue16. This is a structural capacity count, not total
process memory; slices, identifiers, pending journals, model snapshots and the
store have additional costs. No unbounded buffering or state eviction was used.

## Verification

Audited all18 arms: source hashes,64/192 read counts,32/96 write counts,
nearest-rank p99/p95, admitted+dropped=requests, completed labels=50*admitted,
and age sample count=admitted. A private tap enqueue timestamp enables age
measurement without altering packet contents. Three repetitions of frontier,
feedback and temporal-feedback race tests pass; service vet passes.

The frozen v1 load/profiling artifacts remain intact. These outcomes do not
replace their failures. Instrumentation records enqueue-to-update age, excluding
prior Recall assembly. It is not a physical-actuation latency measurement.

## Decision

Do not promote buffering as the throughput rescue. It helps finite bursts but
cannot overcome the measured admission lock bottleneck. A next architectural
candidate must reduce that wait while retaining coherent dependency validation;
simply increasing queue size or dropping snapshot checks is unsupported. Any
change to snapshot publication needs mutation-path and in-progress-commit audits.
Real-task learning, persistent feedback recovery and generation remain separate
open work. No publication, production change, or GitHub push occurred.
