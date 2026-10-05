# Full Recall capacity v24: four workers are nominally over-admitted

Date: 2026-10-01. The [frozen diagnostic](mmm-recall-capacity-v24-protocol.md)
reused the full 200-event Recall fixture and the unchanged v22/v23 store
variants. Each writer trial completed 192 offers, 256 concurrent future-only
writes, exact 200-event nomination/as-of checks, and durable journal reopen.
All nine trials completed; `go vet ./internal/service` passed. A separate full
`-race` replay of the v24 diagnostic also passed (473.144 s); its timings
are excluded from the performance table. This is a
capacity diagnosis, **not** a Goal 6 success or a production change.

The ticker nominally offers one Recall every 8 ms, or 125/s. Four workers
would need mean occupied call time below 32 ms to sustain that nominal rate.
The table reports the range across three rotated writer trials, not pooled
confidence intervals. Queue quarter medians are in **completion order**, not
exact offered-index order.

| Arm | Mean call | Nominal load ratio | Queue p99 | First to last completed-quarter queue p50 |
| --- | ---: | ---: | ---: | ---: |
| Single guarded SQLite | 33.36-33.57 ms | 1.043-1.049 | 103.58-113.36 ms | 0.006-0.008 to 69-93 ms |
| Group, 8 ms dwell | 42.35-43.77 ms | 1.323-1.368 | 491.71-553.91 ms | 12-50 to 390-482 ms |
| Group, 1 ns dwell | 54.98-56.63 ms | 1.718-1.770 | 1083.47-1157.31 ms | 63-69 to 892-958 ms |

Mean journal span was 13.33-13.54 ms for single SQLite, 14.59-15.72 ms
for 8 ms group, and 26.45-28.06 ms for 1 ns group. The single-journal
instrumentation further apportioned 11.59-11.77 ms mean to guard wait and
0.774-0.785 ms to SQLite insert/commit. Internal group guard/insert phases
were **not captured**; the group journal span includes worker queue and
batch scheduling, so these timings do not identify its internal blocker.

All three arms meet the protocol's **nominally capacity-limited** diagnosis:
mean call time exceeds 32 ms and completion-order queue medians grow in each
trial. The single-journal ratio is only about 4-5% above one, so actual
offer-ticker jitter could alter its exact utilization classification. The
larger group ratios and queue growth are less sensitive to that caveat, but
this is still a finite four-worker fixture. Neither p99 alone nor these
means prove which internal lock causes the time. A next experiment should
measure actual offer timestamps and worker occupancy while testing bounded
worker counts or a different publication/admission architecture; it must
retain as-of validation, durable acknowledgement, and the <100 ms loaded
serving gate. Do not relabel the failed v23 candidate as rescued.

Reproduce:

```sh
EVENTFRAME_RUN_RECALL_CAPACITY_V24=1 go test ./internal/service -run '^TestResearchRecallCapacityV24$' -count=1 -v
EVENTFRAME_RUN_RECALL_CAPACITY_V24=1 go test -race ./internal/service -run '^TestResearchRecallCapacityV24$' -count=1
go vet ./internal/service
```

SHA-256 at run: protocol `c3246c25a33f440ad4b37a5baf98042dbfb3884989b00fe28b0ae082557014fd`;
v24 harness `481e34252085fa6eaffe3e894ae6da04651d8f8a037b8eab8cf6274c3e2de0ae`;
v23 helper `4bf91f3301ec462d9e41f4ad17f91d06493915c7a4cba3127182e2429f0ebd61`;
v22 helper `f295522f00333b6a5f408e100a38d4e1a271415d9a60e0dc99e0e98236111e81`.
All seven whole goals remain open.
