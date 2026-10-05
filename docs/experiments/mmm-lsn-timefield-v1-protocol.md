# TimeField availability v1: frozen drop-in rescue probe

The [fractional-time falsifier](mmm-lsn-fractional-availability-v1-results.md)
showed that `available_at: StringField` is not as-of safe. This test-only
probe changes only the private collection's declaration to
`available_at: TimeField`; it keeps the research writer's existing
RFC3339Nano metadata, exact-LSN SQL text, and bound as-of string unchanged.
Write events available at `.1Z` and `.15Z`, then query as-of `.1Z`,
`.12Z`, and `.15Z`. Required sets are `{early}`, `{early}`, and
`{early, late}`; compare with the ordinary time-aware `Store.Search`.
Insertion, binding or comparison failure counts as a failed drop-in
rescue. No value normalization or new metadata field is allowed in this
probe. A pass would still require a later loaded and migration study.
