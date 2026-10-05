# Rolling-LSN reader v2: timing pass, as-of contract fails

Protocol: [mmm-rolling-lsn-reader-v2-protocol.md](mmm-rolling-lsn-reader-v2-protocol.md).
The opt-in test-only comparison used fresh schema-backed private collections,
three rotated pairs per run, 200 past events, 16 future-only events, 256
visible writes, and 192 k=50 read offers per arm. A harness audit corrected
the lag metric to count only reads **offered after** each write acknowledgment.
The earlier sub-millisecond lag figures from the uncorrected run are
superseded.

Two uninstrumented runs of the corrected code passed the frozen
*whole-second fixture* gates:

| Run | Current / rolling pooled read-call p99 | Ratio | Writer completion ratio | Current / rolling post-ack lag p99 |
| --- | --- | --- | --- | --- |
| 1 | 12.102 / 10.777 ms | 0.891 | 0.995 | 14.900 / 15.564 ms |
| 2 | 13.402 / 12.091 ms | 0.902 | 0.997 | 18.009 / 14.881 ms |

Each run had 576 correct reads per arm, 256 acknowledged writes per arm,
per-version motion, no future IDs in those whole-second fixtures, a final
read containing batch 15, and zero active temporal leases. Retained bytes
reported by the temporal API were 10,064 per arm; this is not RSS or a
steady-state memory measure. The race-instrumented run reported no data race
but **failed the timing gate** under instrumentation (pooled read-call p99
240.000 ms), so it is not an instrumented pass and its timings are excluded
from the uninstrumented performance claim.

**Overall decision: not an as-of-safe design.** A follow-up
[fractional-time falsifier](mmm-lsn-fractional-availability-v1-results.md)
showed that the same `StringField` SQL predicate leaks an event available
at `.15Z` into a `.1Z` read and also omits past events. The frozen v2
numbers demonstrate a possible reader/writer timing opportunity only on
the narrow whole-second fixture. They cannot justify adoption, a Goal 6
pass, or a general no-future-data claim. A typed or fixed-width availability
key and its migration/validation path must pass before loaded confirmation.
Production remains untouched.

Reproduce the corrected timing screen with:

```sh
EVENTFRAME_RUN_ROLLING_LSN_READER_V2=1 go test ./internal/store/libravdbstore -run '^TestResearchRollingLSNReaderV2$' -count=1 -v -timeout 5m
```

Source SHA-256 at corrected runs: test `8bc833c6011a53cef18a94225ff79a3120213ef28e50fc05cda1f042b23a7856`;
protocol `e76d4e439ebf2675b0dcf8768a92b985f477830eb42291d868483cbbb11964ce`.
