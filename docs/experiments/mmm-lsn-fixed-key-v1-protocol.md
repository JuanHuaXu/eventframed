# Fixed-width availability key v1: frozen SQL contract probe

StringField and TimeField over the existing variable-width `available_at`
metadata both failed the [fractional as-of falsifier](mmm-lsn-fractional-availability-v1-results.md).
This isolated test uses a **new** private metadata key, `available_at_sort`,
with a fixed-width UTC format `2006-01-02T15:04:05.000000000Z`. The key
must be derived from the same `time.Time` used for the EventFrame, not
supplied independently by a caller. This is a storage-design probe only;
the existing writer, production collection and migration are unchanged.

Insert `early` at `.1Z` and `late` at `.15Z` with both their ordinary
variable-width `available_at` and fixed-width `available_at_sort` metadata.
At an exact durable LSN, query with
`WHERE available_at_sort <= $available_by_sort` at `.1Z`, `.12Z`, and
`.15Z`. Expected sets are `{early}`, `{early}`, and `{early, late}`.
Close/reopen and repeat all three. A different-zone `time.Time` denoting
the same instant must generate the identical UTC sort key. The test fails
on any wrong ID, insert/bind error or reopen difference. A pass establishes
only this finite SQL contract; ingestion integration, general property
tests, old-row backfill, loaded latency and Goal 6 remain separate.
