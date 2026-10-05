# Origin-binding adapter: repeat rescue with a fallback limitation

[Protocol](ORIGIN_BINDING_PROTOCOL.md), [public outputs](public-provenance-bound.json),
[adapter](../originbinding/binding.go), [tests](../originbinding/binding_test.go),
[fallback diagnostic](../originbinding/fallback_test.go).

The offline registry adapter binds exact serialized event payloads to stable
origins supplied by the fixture controller. It changes only the candidate's
packet correlation key. It does not create an authenticated source registry,
alter event data or confidence laws, or grant authority to claimed URLs.

## Public fixture results

All52 paired repeats are suppressed once each by the actual default packet
selector. Stable-origin pairs remain13/13, while missing-origin, producer-change
and per-fetch-origin pairs improve from0/13 to13/13 each. Each pair occupies one
slot instead of two. This is deterministic behavior on13 consumed NASA facts,
not52 independent generalization samples or a measured source-reliability rate.

The registry is supplied using the known fixture source mapping. Both receipts
are constructed outside the adapter and bind the entire observed payload. This
tests what happens GIVEN correct trusted origin bindings. Establishing those
bindings safely in a real capture pipeline remains open.

The output's `correlated` column deliberately retains the original descriptor
comparison; `same_group` and packing counts reflect the adapter. The explicit
bound group can therefore suppress a pair the old descriptor does not recognize.

## Negative checks and residual gap

Missing records, altered payloads, wrong tenants, unsupported kinds, duplicate
identities, empty origins and registry overflow are rejected without modifying
the candidate. Caller mutation of the input binding slice does not modify the
registry. Certified AP split buckets are preserved by packing. Distinct exact
relations/origins with separate lineages stay separate; explicitly declared
observed occurrences also stay separate under the same producer/lineage.

However, the widened same-lineage check exposes a residual integration gap:
`A > B` and `A < B` get different bound keys but packing's legacy correlation
fallback normalizes away the operators and suppresses one. The earlier different-
producer checks did not cover this case. The diagnostic test explicitly records
the known limitation; its passing status is NOT a successful separation test.

Therefore this is a repeat-suppression rescue only, not a fully validated
provenance-aware packing implementation. No broad core patch was made. The next
research change must reconcile bound-key evidence with legacy fallback while
preserving unbound behavior, genuine repeats and AP authority. It must not use
fake AP certificates to bypass the fallback or trust attacker-created bindings.

## Verification and scope

Public output replays byte-for-byte. Six focused test functions pass under the
race detector, including the explicitly labeled limitation diagnostic. Runtime
sources are unchanged. The adapter is capped at256 offline entries but hashes
full events; no hot-path performance claim is made. This is not full ingestion,
signature verification, posterior updating or an actual-agent experiment.
All seven whole research directions remain open.
