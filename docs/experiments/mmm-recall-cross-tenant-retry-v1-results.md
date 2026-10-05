# Cross-tenant journal retry v1: confirmed global invalidation

Date: 2026-10-02. Frozen [protocol](mmm-recall-cross-tenant-retry-v1-protocol.md).
The real `Service.Recall` queried tenant A. Immediately before its first
frontier-journal write, the test ingested one event for tenant B. Both
memory and LibraVDB stores behaved identically:

| Tenant B write | Journal attempts | Packet/journal runtime version | Tenant A candidates |
| --- | ---: | ---: | --- |
| Future to `as_of` | 1 | 2 / 2 | A seed only |
| Visible at `as_of` | 2 | 3 / 3 | A seed only |

The visible B event changed no A candidate but forced a retry under the
global `JournalSnapshotCompatible` rule. The future B event advanced the
Store after capture without affecting the as-of decision, so the journal
accepted the older snapshot. No B event entered A's packet. Focused normal
and race tests, ordinary service tests and vet pass.

This isolates one source of cross-tenant retry pressure; it does not prove
that tenant B can always be ignored. Graph, policy, certificates, provenance
and other cross-tenant dependencies would need explicit scope before a
tenant-aware compatibility guard could replace the global rule. The
[concurrent screen](mmm-recall-presearch-concurrent-v1-results.md) still
fails completion and latency gates even though the pin prevents observed
silent omissions. Production was untouched; Goal 6 remains open.

Reproduce:

```sh
EVENTFRAME_RUN_RECALL_CROSS_TENANT_RETRY_V1=1 go test ./internal/service -run '^TestResearchRecallCrossTenantRetryV1$' -count=1 -v -timeout 5m
EVENTFRAME_RUN_RECALL_CROSS_TENANT_RETRY_V1=1 go test -race ./internal/service -run '^TestResearchRecallCrossTenantRetryV1$' -count=1 -timeout 5m
```

SHA-256: test
`f16bea0740f8ac11d06f5bf42c36407abc10a5dbe3be2994a17033f6db7b39fd`;
protocol `a77a78b689dab4bad7ec58749d5e81912e635e54dac0d2ffe6a6f69e3728b79a`.
