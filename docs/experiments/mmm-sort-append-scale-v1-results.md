# Incremental publication append scale v1: isolated result

**Subsequent audit:** This is a pre-interleaving-guard measurement. The
[batch publication follow-up](mmm-sort-batch-publication-v1-results.md)
records the final gate invariant and paired batch-versus-single costs.

Date: 2026-10-02. Protocol: [v1](mmm-sort-append-scale-v1-protocol.md).
The opt-in test passed its correctness checks in 44.20 s. Both private Stores
received the same events and had equal row counts at the start of each
measured pair. One Store used the sortable receipt writer as control; the
other used the receipt-bound journal and READY marker. Each block contains
100 rotated-order write pairs and 100 quiet candidate captures.

| Rows before block | Rows after block | Control write p50 / p99 / max | Journaled write p50 / p99 / max | Candidate/control p50 / p99 | Capture p50 / p99 |
| ---: | ---: | :--- | :--- | :--- | :--- |
| 35 | 135 | 6.852 / 7.941 / 8.466 ms | 7.100 / 8.233 / 8.912 ms | 1.036 / 1.037 | 5.583 / 7.958 us |
| 259 | 359 | 11.977 / 14.952 / 16.064 ms | 12.314 / 14.979 / 16.101 ms | 1.028 / 1.002 | 5.541 / 13.667 us |
| 1027 | 1127 | 32.982 / 41.898 / 43.023 ms | 33.201 / 40.194 / 40.254 ms | 1.007 / 0.959 | 5.625 / 17.833 us |

The journal adds a small measured median increment in these fixtures, while
the raw one-event writer rises sharply with collection size. The 1,027-row
p99 ratio below one is **not** evidence that journaled writes are faster:
there was only one 100-pair block per size, and the arms used separate
Stores. The dominant isolated cost is the underlying one-event write, not
the incremental SQLite journal. This makes the current single-event path an
unpromising route to sustained 4 ms offered arrivals on one serial owner;
that is a design inference, not a measured loaded-service failure for this
candidate. A genuinely different batch/admission/publication architecture
must retain per-event acknowledgement and as-of safety, then be tested with
concurrent Recall and labels. Do not simply loosen the freshness target.

This test did not measure loaded Recall, writer queue time, crash recovery,
multi-owner safety, or large-corpus scaling. Production was untouched.
Goal 6 and all seven whole goals remain open.

Reproduce:

```sh
EVENTFRAME_RUN_SORT_APPEND_SCALE_V1=1 go test ./internal/store/libravdbstore -run '^TestResearchIncrementalSortAppendScaleV1$' -count=1 -v -timeout 5m
go test ./internal/store/libravdbstore -count=1
go vet ./internal/store/libravdbstore
```

The focused command, ordinary package tests and vet passed. Source SHA-256:
test `8a8fd9e17a211d43f45674f0e310e5cd2361c390d2cab6db5dd17fa8f03f69a0`;
protocol `3fa38feadaa85b4c52a392654f4fccde630b3de928bbe23d5cb910509a6e604a`.
