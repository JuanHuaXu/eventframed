# Pre-Timing Ownership Addendum

V3's first ordinary/race preflight repaired the profile adapter. Before cost
source freeze, readback of Collection.Search found that it takes ownership of
nonempty index-result vectors without cloning. The immutable prototype base
has no provider, so HNSW returns a borrowed raw-store view. This risks caller
mutation, use after retirement, and returning normalized index vectors instead
of authoritative unnormalized record vectors. Do not disguise this as a profile
fix or claim V3 has only the one new change in its initial protocol.

Add two independent lifecycle regressions: mutation of a returned standalone
vector must not change later predictions; a provider-backed collection result
must omit index vectors so storage hydrates its owned canonical record. Preserve
pre-patch failures and fix before frozen cost measurement. Standalone results
clone the generation's canonical index vectors; provider-backed results remain
nil at the index boundary. The immutable graph never consults mutable storage.
No benchmark dimensions, caps, quality gates or scientific thresholds changed.
