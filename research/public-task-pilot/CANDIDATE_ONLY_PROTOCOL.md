# Private immutable candidate-only retrieval

Research overlay only, no module-cache or public search modification. Add an
explicit ResearchImmutableCandidates method restricted to HNSW/cosine collections.
It reuses the same locked nomination, shard collection, normalization and top-k
logic but omits payload hydration. Fail on ordinal-only/empty-ID candidates.
Existing Search and Query execute the original hydrated path unchanged.

Only the private HNSWBase adapter calls the new method. Its base map owns complete
vectors, validates every returned ID and rescores before delta merge. The method
itself does not establish immutability; private derived-index ownership supplies
that contract. Never use it as a replacement for mutable public record retrieval.

First compare same-graph candidate sets/scores against ordinary search and verify
ordinary vectors/metadata/version still return. Run race/lifecycle tests. Then
repeat the partition static exact-oracle screen, followed by the unchanged async
retirement load if equivalence holds. Preserve failed results. No changed effort,
delta/lease limits, durability, arrival rates or acceptance gates.
