# TimeField availability v1: drop-in rescue fails

Protocol: [mmm-lsn-timefield-v1-protocol.md](mmm-lsn-timefield-v1-protocol.md).
A private collection declared `available_at: TimeField`, but the research
writer and bound as-of strings were unchanged. Insertion and SQL binding
succeeded. The exact-LSN SQL returned `{early, late}` at `.1Z`, `{}` at
`.12Z`, and `{late}` at `.15Z`; the correct ordinary `Store.Search` sets
were `{early}`, `{early}`, and `{early, late}`. These are identical to the
StringField failure, including future leakage at `.1Z`. Merely changing
the schema type does not make the existing string values time-ordered in
this query path.

The failed test is intentionally retained. A later candidate must store a
separately validated sortable key or use a proven timestamp conversion,
then address existing rows before any serving switch. No production change.

```sh
EVENTFRAME_RUN_LSN_TIMEFIELD_V1=1 go test ./internal/store/libravdbstore -run '^TestResearchLSNTimeFieldV1$' -count=1 -v -timeout 2m
```

Source SHA-256 at run: test `a7a0b3e319ed60ecf24770ed2cb4b2d76a31fc6e6d2218989a75293ca4208a57`;
protocol `32a079a5c21745780cb73b11d9bc17f125cc53a311f9b4d76df8ec4a086347c0`.
