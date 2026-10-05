# Query-scoped posterior validity v20: conditional component pass

Date: 2026-10-02. The frozen [protocol](mmm-posterior-dependency-v20-protocol.md)
passes its **pure guard** tests. It is not a Goal 6 rescue and is not wired
to the Store, Recall, or scored law. All seven whole goals remain open;
production is untouched.

## What was checked

`internal/researchvalidity.Check` accepts a retained posterior only when
tenant, posterior key, query, model, horizon, source and frontier digest
match; semantic versions remain stable; and a contiguous bounded mutation
witness explains every runtime and evidence-epoch transition. A future-only
insert is harmless for an as-of request. A visible insert is harmless only
with an independently certified score upper bound strictly below the
top-k cutoff and a complete non-impacting graph dependency check. An
available related outcome blocks reuse. Missing score/dependency coverage,
epoch accounting, availability, or a runtime step returns `Unknown`.

Focused cases pass for future-only and below-cutoff visible appends,
equal-cutoff competition, graph impact, related/unrelated outcomes,
future related feedback, stale source, changed query/frontier/graph,
unclassified writes and missing epoch/runtime steps. Harmless prefixes
remain compatible through 128 mutations; adding one competing insert
blocks. A v19-like high-relevance insert with score upper bound
`cos(.00213)` against cutoff `cos(.74)` blocks, as it should. This last
check is a constructed frontier-churn case, not a replay of v19's Store.

`go test -race ./internal/researchvalidity -count=1` and
`go vet ./internal/researchvalidity` pass. A three-run microbenchmark on
Apple M4 measured 230.4-230.8 ns/op for 128 certificate mutations and
19,598-19,814 ns/op for 10,000, with zero measured allocations. These
times cover only an in-memory scan of supplied witnesses. They exclude
durable log construction, provenance proofs, score bounds, graph checks,
retrieval, publication, and service contention.

## Decision and next step

The epoch-accounting audit found and closed a false-compatible path in the
first draft: a certificate publication could have concealed an unexplained
`EvidenceEpoch` increment. The final code requires each mutation to report
the resulting epoch and verifies the allowed transition.

The current service does not retain an independently attested exact
frontier digest, ranking-space upper bounds, or complete mutation/dependency
provenance. It also keys some posteriors by event rather than by the full
query/selection regime. Therefore this guard cannot yet be used to keep
old posteriors scored after visible writes. V19's high-relevance churn is
expected to block, and its zero-scored-learning result is unchanged.
The next viable Goal 6 question is whether a durable, bounded as-of
provenance index can produce these witnesses and valid certificates cheaply
enough under loaded full Recall; no production change follows from v20.
