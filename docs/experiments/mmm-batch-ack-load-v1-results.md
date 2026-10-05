# Batch acknowledgement load v1: reader-tail screen fails

The [frozen protocol](mmm-batch-ack-load-v1-protocol.md) **FAILED** its
backend contention screen. A test-only authority-protected queue delivered
each of 256 future-only writes only after the corresponding raw LibraVDB
batch and SQLite sidecar finalized. Across three rotated fresh pairs, all
six arms accepted 256 unique writes, completed 192 concurrent past-as-of
vector searches, preserved every per-version motion entry, and returned no
future event. These are backend searches, not full `Service.Recall` calls.

| Pair | Arm | Batches | Ack p99 | Search-call p99 | Writer completion | Write offer gap p50 |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| 0 | single | 256 | 999.865 ms | 12.139 ms | 3.269 s | 11.111 ms |
| 0 | batch-16 | 17 | 34.731 ms | 17.108 ms | 0.278 s | 1.002 ms |
| 1 | batch-16 | 18 | 46.135 ms | 19.523 ms | 0.301 s | 1.000 ms |
| 1 | single | 256 | 979.070 ms | 11.988 ms | 3.269 s | 11.221 ms |
| 2 | single | 256 | 1006.791 ms | 13.483 ms | 3.305 s | 11.308 ms |
| 2 | batch-16 | 17 | 42.722 ms | 19.341 ms | 0.286 s | 0.999 ms |

Pooled acknowledgement p99 was **999.865 -> 42.722 ms** (ratio 0.043),
and summed writer completion was **9.842 -> 0.865 s** (ratio 0.088).
These pass the frozen writer criteria. Pooled search-call p99 was
**12.198 -> 18.755 ms** (ratio **1.538**), failing the <=1.10 reader
criterion. Each pair has the same direction of reader regression. Batch
sizes were mostly 15 or 16: the candidate used 17, 18 and 17 batches
respectively; all three controls used 256 single writes.

The single-write producer could not maintain its nominal 1 ms offer rate:
its median actual offer gap was 11.1-11.3 ms, while batch-16 sustained
about 1 ms. Thus the arms faced the same attempted work but not the same
realized overlapping write rate. The candidate's reader harm may be caused
by longer index/transaction exclusion, higher simultaneous write pressure,
CPU contention, or a combination; this test does not isolate root cause.
The writer-tail speedup must not be presented as a matched-realized-rate
service gain. No result here meets the unchanged full-Recall/learner
4 ms Goal 6 gate.

The first `-race` attempt exceeded a 30-second fixture context under
instrumentation; extending only that test execution ceiling allowed all six
arms to finish. That run reported no data race, but exited nonzero because
the same frozen reader-tail gate failed. Its instrumented timings are not
used above. Ordinary package `go test -race` (with the opt-in load screen
disabled) and `go vet` pass. Preserve the negative batch-16 result; a new
candidate needs a frozen design that protects reader tails, not merely a
larger or faster write batch. All seven whole goals remain open and
production is untouched.

Reproduce the expected-failure screen:

```sh
EVENTFRAME_RUN_BATCH_ACK_LOAD_V1=1 go test ./internal/researchbatch -run '^TestBatchIntentAckLoadV1$' -count=1 -v -timeout 5m
go test -race ./internal/researchbatch -count=1 -timeout 3m
go vet ./internal/researchbatch
```

At-run SHA-256:

```text
4ce74b342faf48dff70191863466f1b7f47b8ef78f5330e461c0ac6f9fc43dee  internal/researchbatch/batch_ack_load_test.go
ef59fe430c8a59f42a8016b2a0e24d72395f32e28f7f3324e3afe3818804ccc0  docs/experiments/mmm-batch-ack-load-v1-protocol.md
c339ff6c985202fc9dcbde7a9280677ce290285125148f9e6ab35b8c172f66fe  internal/researchbatch/batch_intent_prototype_test.go
```
