# Rolling-LSN reader v1: blocked design screen

Protocol: [mmm-rolling-lsn-reader-v1-protocol.md](mmm-rolling-lsn-reader-v1-protocol.md).
The frozen three-pair visible-write comparison did **not** complete and
cannot supply a reader or freshness latency ratio. Preserve both failures:

1. The first current-path control arm exhausted its four permitted
   publication attempts while visible batches were arriving. The rejected
   captured runtime versions were 408, 424, 440 and 456; the latest version
   at failure was 472. The post-read compatibility check invalidated every
   attempt. This is evidence of a retry-storm boundary in the strict
   post-read rule at this offered rate, not evidence that the rolling SQL
   path is slower or faster.
2. A separately labeled candidate-only diagnostic failed in warm-up before
   any timed offer: `bind error: identifier 'available_at' not found in scope
   or catalog`. The current event collection is created without a metadata
   schema for `available_at`; its stored metadata alone does not make the
   field a SQL predicate in this LibraVDB release. Removing the predicate
   would permit future-available records and invalidate the as-of contract.

The earlier fixed-LSN future-only result remains a valid *component* screen,
but neither failure here was retuned away. The next design must separately
establish (a) a real as-of filter, such as a schema-backed availability field
with migration/reopen controls, and (b) a coherent read linearization rule.
An exact captured LSN can in principle linearize a read before a concurrent
commit; whether the service freshness contract permits that must be frozen
and tested, not assumed. A post-read current-version requirement is stronger
and can starve under frequent visible writes. Any replacement comparison
must retain negative future-row controls and measure capture, pin, query,
release, retries, and write interference. No production path changed.

Reproduce the two diagnostic failures with:

```sh
EVENTFRAME_RUN_ROLLING_LSN_READER_V1=1 go test ./internal/store/libravdbstore -run '^TestResearchRollingLSNReaderV1$' -count=1 -v -timeout 5m
EVENTFRAME_RUN_ROLLING_LSN_DIAGNOSTIC=1 go test ./internal/store/libravdbstore -run '^TestResearchRollingLSNCandidateDiagnostic$' -count=1 -v -timeout 5m
```

Source SHA-256 at run: test `4297d7a67d9027882b3dcd0214ccab2b6bd64bbd4f2c3ad7f41a6b191136a405`;
protocol `9ef48584a8ef5dd23f7dcb168269e3203e326ae692e49cd5425d5bd92e33500e`.
