# Fractional availability: frozen as-of falsifier

The private schema and rolling-LSN v2 screens used whole-second timestamps.
`research_event_batch.go` formats `available_at` with a Go layout containing
nine `9` fractional digits, which can omit trailing zeros. If LibraVDB SQL
compares `StringField` values lexically, `.1Z` may sort *after* `.12Z`.

In a private schema-backed collection, write two events available within
one second: `early` at 2026-10-02T00:00:00.1Z and `late` at
2026-10-02T00:00:00.15Z. Query exact latest LSN using the unchanged
`available_at <= $available_by` predicate with as-of
2026-10-02T00:00:00.12Z. Correct result is `{early}`. Also query as-of
`.1Z`, where correct result is still `{early}` and `late` would be
unwitting future data. Compare with `Store.Search` at the same as-of.
Then query as-of `.15Z`; correct result is `{early, late}`. The `.1Z`
case was added after the first diagnostic failure and is an additional
falsifier, not a replacement cohort. No postfilter or string-format change is allowed in
this falsifier. If SQL differs, the earlier whole-second passes remain
narrow measurements but not a general as-of-safe design.
