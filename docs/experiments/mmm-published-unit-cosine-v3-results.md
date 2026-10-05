# Published-LSN unit-vector cosine v3: restricted component passes

Date: 2026-10-02. Frozen [protocol](mmm-published-unit-cosine-v3-protocol.md).
On a private declared-payload collection whose nonzero stored vectors and
query were checked unit-normalized, SQL `<=>` plus
`clamp(1-distance,-1,1)` matched ordinary Store Search for all five
eligible rows within `1e-5`:

| Contrast | SQL distance | Converted and Store similarity |
| --- | ---: | ---: |
| Two aligned rows | 0 | 1 |
| Angled `(0.6,0.8)` | 0.399999976 | 0.600000024 |
| Orthogonal | 1 | 0 |
| Opposite | 2 | -1 |

All EventFrame bodies decoded, SQL ordering was descending in converted
similarity, and the future sentinel was excluded. The test-only validator
rejected a scaled query `(2,0,0,0)`, a nonunit stored vector `(3,4,0,0)`
and zero. Focused normal/race tests, ordinary package tests and vet pass.

This is a **restricted score component**, not a general SQL Search
replacement. No production writer, pre-embedded Observe path, query
boundary, or migration enforces unit norm today. Existing nonunit records
must be detected or migrated, and out-of-contract inputs must fail closed
or be normalized under a declared representation rule. The v2 failure
remains. Full Recall read-view coherence, journal routing, performance and
production are still open; Goal 6 is not complete.

Reproduce:

```sh
EVENTFRAME_RUN_PUBLISHED_UNIT_COSINE_V3=1 go test ./internal/store/libravdbstore -run '^TestResearchPublishedUnitCosineV3$' -count=1 -v -timeout 5m
EVENTFRAME_RUN_PUBLISHED_UNIT_COSINE_V3=1 go test -race ./internal/store/libravdbstore -run '^TestResearchPublishedUnitCosineV3$' -count=1 -timeout 5m
go test ./internal/store/libravdbstore -count=1 -timeout 5m
go vet ./internal/store/libravdbstore
```

SHA-256: test `c9dc52af189c4c4796eff68ceeebdc684d508ec5de0f481db8f4b5f47c202edd`;
protocol `a6492755d0c7c644eb698ef9c6fd9aae8fbbc7b73eaac7f4fb5933944e89e44f`.
