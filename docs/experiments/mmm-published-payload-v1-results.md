# Published-LSN payload projection v1: declared-schema component passes

Date: 2026-10-02. Frozen [protocol](mmm-published-payload-v1-protocol.md).
The existing availability-only research collection cannot project
`event_json`: SQL returns `bind error: identifier 'event_json' not found in
scope or catalog`. The same private fixture with `event_json`, `corpus_text`
and `raw_content` declared as metadata string fields passes.

At certified LSN21, SQL returned and decoded exactly `past100` and `at120`
at as-of `.120Z`. After a gate-authorized eligible append, LSN25 returned
those two plus `new110`; querying LSN21 still excluded `new110`. Both LSNs
excluded the `.125Z` future event. The result bodies passed the same
`decodeStoredEvent` validator used by Store Search, including raw content,
identity and canonical corpus. The new-LSN body query survived close/reopen.
Focused normal/race checks, ordinary package tests and vet pass.

This establishes a **declared-schema body projection**, not parity of
retrieval scores. The separate [score probes](mmm-published-score-v1-results.md)
show why the SQL `SearchResult.Score` cannot be copied into full Recall.
Existing collections need migration/backfill and readiness checks; this
test does not supply them for payload columns. Graph, posterior, journal,
full Service Recall, loaded latency and production remain untouched.

Reproduce:

```sh
EVENTFRAME_RUN_PUBLISHED_PAYLOAD_V1=1 go test ./internal/store/libravdbstore -run '^TestResearchPublishedPayloadV1$' -count=1 -v -timeout 5m
EVENTFRAME_RUN_PUBLISHED_PAYLOAD_V1=1 go test -race ./internal/store/libravdbstore -run '^TestResearchPublishedPayloadV1$' -count=1 -timeout 5m
```

SHA-256: test `4b0c2c8d48863165c6f36b4b671e756fac1147a90cf8d869907625497ab775cf`;
protocol `4bbcdadfd6b5095e2a1882195375a40300c17e70934a2dd8c8f83efc675c9c46`.
