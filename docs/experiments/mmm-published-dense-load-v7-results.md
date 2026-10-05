# Dense 256D published-LSN loaded read v7: split outcome

Date: 2026-10-02. The [frozen protocol](mmm-published-dense-load-v7-protocol.md)
was run on fresh private 256D collections in cap order 50, 200, 200,
50. Each trial offered 160 eligible writes and 160 reads at nominal
4 ms cadence, with eight reader workers and a single cap-16/16-ms-dwell
journaled writer. At the start each corpus contained 200 eligible and
17 future rows. Every read decoded full EventFrames and was checked
against an independent top-k oracle indexed by the exact published LSN.

Two uninstrumented executions passed all four frozen trials each:
1,280/1,280 reads matched LSN-specific membership and score checks,
1,280/1,280 writes were acknowledged, no future or unpublished row
appeared, and there were no read or writer errors. Across the four
trials, the first/second execution's worst p99 and maximum were:

| Metric | First normal run | Second normal run | Frozen limit |
| --- | ---: | ---: | ---: |
| Write offer-to-ack p99 | 30.539 ms | 31.160 ms | <250 ms |
| Read-call p99 | 5.194 ms | 5.195 ms | <100 ms |
| Read offer-to-done p99 | 5.195 ms | 5.196 ms | <100 ms |
| Published-view maximum age | 28.410 ms | 28.177 ms | <250 ms |

Actual write/read offer-gap medians stayed near 4 ms. Each normal
trial formed 36-38 batches for 160 writes, not a synthetic sequence of
ten full batches. The read and write clocks overlapped.

The `-race` run completed the same four trials without a reported data
race, read/write error, oracle mismatch, or future leak. **It failed the
frozen timing gate in all four trials**: write-age p99 ranged from
1.032 to 1.295 seconds, with only 11 batches per trial as the
instrumented writer fell behind the fixed offer schedule. Read-call
p99 stayed below 44 ms and view age below 228 ms. This is evidence
of a race-instrumented throughput limit, not evidence that ordinary
serving misses the same limit; equally, the frozen all-mode screen
cannot be recorded as an unconditional pass. No thresholds, schedule,
or vectors were retuned after observing it.

Ordinary package tests and vet pass. The result advances a loaded
backend component under the single-owner/no-bypass contract. It does
not measure full Service Recall, Bayesian frontier journal writes,
learning freshness from a label to a served forecast, network or
OpenClaw latency, large-corpus behavior, or agent outcomes. Goal 6
and all seven whole goals remain open; production was untouched.

Reproduce:

```sh
EVENTFRAME_RUN_PUBLISHED_DENSE_LOAD_V7=1 go test ./internal/store/libravdbstore -run '^TestResearchPublishedDenseLoadV7$' -count=1 -v -timeout 5m
EVENTFRAME_RUN_PUBLISHED_DENSE_LOAD_V7=1 go test -race ./internal/store/libravdbstore -run '^TestResearchPublishedDenseLoadV7$' -count=1 -timeout 5m
go test ./internal/store/libravdbstore -count=1 -timeout 5m
go vet ./internal/store/libravdbstore
```

SHA-256: test `a2df225e663aa86ddd3d8b125ac3dd4e96d847fdaa3c17385aa234d43f392568`;
protocol `d35e7c3d8e87536183a088d0c29bb45583db0cd88f5df5ee586b762b43d9534f`.
