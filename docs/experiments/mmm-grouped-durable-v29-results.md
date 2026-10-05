# Grouped durable admission v29

Status: durable age/availability rescue FAILED. Integrity/accounting PASS.

Added actual Durable SQLite admission, complete record readback comparison and
explicit unlabeled discard inside the existing guarded group callback. Group4
without persistence and off were rerun as same-execution controls. Small persistent
read-only/concurrent-write scenarios passed three race repetitions; vet passed. A separate
cold-preview abandonment test verifies contiguous durable IDs, complete record
preservation after reopen, no feedback after discard and continued admission.

| Trial | Mode | Accepted / 192 | Queue drops | Read p99 (ms) | Write p99 (ms) | Age p95 (ms) |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| 0 | Off | N/A | N/A | 32.968 | 19.020 | N/A |
| 0 | Group4 | 192 | 0 | 24.893 | 24.518 | 107.301 |
| 0 | Durable | 78 | 114 | 20.681 | 151.364 | 857.593 |
| 1 | Off | N/A | N/A | 30.972 | 18.938 | N/A |
| 1 | Group4 | 192 | 0 | 21.042 | 19.978 | 117.798 |
| 1 | Durable | 77 | 115 | 25.008 | 151.943 | 788.246 |
| 2 | Off | N/A | N/A | 36.900 | 21.337 | N/A |
| 2 | Group4 | 192 | 0 | 27.906 | 22.909 | 155.112 |
| 2 | Durable | 77 | 115 | 24.964 | 167.602 | 830.970 |

Nearest-rank quantiles. Durable admitted 232/576 (40.28%) and dropped 344. All
11,600 admitted candidates had actual admission and discard writes with complete
original-record comparison; zero errors, stale rejections or entry expiries.
There were no labels or fitting. Stored discards are not negative evidence.

All durable age tails exceed 250ms. Lower read p99 does not rescue this outcome:
writes are delayed about eightfold and most observations are dropped. Total
callback time is 992/906/949ms across the three durable arms, versus only
109/110/114ms entering guards. The bottleneck has returned to callback work.
This experiment includes both write APIs and integrity readback, so it does not
yet attribute the cost specifically to COMMIT/fsync.

Next instrument admission, readback and discard API time separately before
implementing bounded group transactions. Preserve FULL synchronous durability,
original record semantics, exact retries and explicit terminal types. Do not
remove integrity checks or claim loaded continuous learning from this fixture.

## Artifact

`mmm-grouped-durable-v29.jsonl`, SHA-256
`ab6dddd76b6f92c0db694a37c18ccd7d4e6e281636133f63dedb877596e8a6df`.
Independent parsing verified nine distinct cells, embedded hashes, 192 reads/
96 writes per arm, group-weighted outcome conservation and actual
admits=discards=validated=50*accepted in durable arms. This verifies accounting,
not performance acceptance or cross-store atomicity. No deployment or push.
