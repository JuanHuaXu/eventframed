# Fractional availability v1: as-of failure

Protocol: [mmm-lsn-fractional-availability-v1-protocol.md](mmm-lsn-fractional-availability-v1-protocol.md).
The unchanged research writer stores `available_at` as a variable-width
RFC3339Nano-style string. An exact-LSN SQL query on a private collection
with `available_at: StringField` used the same `<=` predicate as the
rolling-LSN screens. The current time-aware `Store.Search` was the control.

| As-of | Correct / current Search | SQL result | Defect |
| --- | --- | --- | --- |
| `00:00:00.1Z` | `{early}` | `{early, late}` | **Future leak**: `late` is at `.15Z` |
| `00:00:00.12Z` | `{early}` | `{}` | False omission |
| `00:00:00.15Z` | `{early, late}` | `{late}` | False omission |

This confirms that declaring the field as `StringField` does not provide a
general time-ordering contract. Variable-width fractional strings are not
lexicographically ordered like instants. The earlier whole-second schema
probe and rolling-LSN v2 timings remain valid *for their fixtures* but do
not qualify this SQL path for serving arbitrary as-of requests. In
particular, the v2 read-latency gate cannot override a future-data leak.

Next falsifier: check whether LibraVDB `TimeField` interprets the unchanged
stored RFC3339Nano strings as instants. If it does not, a separately
versioned, fixed-width or numeric availability key plus migration controls
will be needed. No production data or schema was changed.

Reproduce with:

```sh
EVENTFRAME_RUN_LSN_FRACTIONAL_AVAILABILITY_V1=1 go test ./internal/store/libravdbstore -run '^TestResearchLSNFractionalAvailabilityV1$' -count=1 -v -timeout 2m
```

Source SHA-256 at run: test `cd81105356ee7715af50828d8dd1a35e64528ff27ab82c8aed917f0ea4705b2d`;
protocol `27e552cf3878cd5d0909cb1be6016fc770b5cdd87a6aad9cfeba0d2c2c2c9581`.
