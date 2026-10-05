# Published-LSN direct score v1: failed contract

Date: 2026-10-02. Frozen [protocol](mmm-published-score-v1-protocol.md).
With an aligned and an orthogonal visible vector, SQL `<->` returned
projected distances 0 and 2 and placed those values in its
`SearchResult.Score`. Ordinary cosine `Store.Search` returned similarities
1 and 0. The bodies and aligned-before-orthogonal ordering were correct,
but copying the SQL score would invert the ranking scale. The opt-in test
deliberately remains **failing**.

The local LibraVDB v1.6.13 optimizer defines SQL `<->` as L2 distance
independent of the collection metric; ordinary cosine Search converts
its own metric distance into higher-is-better similarity. An explicit
operator and conversion are required. The [cosine probe](mmm-published-cosine-v2-results.md)
tests the next candidate. No full Service or production change follows.

Reproduce the expected failure:

```sh
EVENTFRAME_RUN_PUBLISHED_SCORE_V1=1 go test ./internal/store/libravdbstore -run '^TestResearchPublishedScoreV1$' -count=1 -v -timeout 5m
```

SHA-256: test `f5cee432adc221e6fc7b72e08164cccf9dd130004eab2c52f09aa2761e07599d`;
protocol `ad43b2e5a5fca6502dea40aea4541971646ce139a8cd01b42b56140e4cd29d1b`.
