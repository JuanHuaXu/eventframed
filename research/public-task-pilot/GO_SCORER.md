# Frozen sparse scorer Go integration

internal/researchsparse matches the frozen Python pilot on160 public candidate
rows. Export fitting uses design only; this is parity evidence, not fresh
confirmation. The immutable model copies264 weights and rejects nonfinite,
oversized or uninitialized inputs. Features read six compressed fields only,
not content, IDs or attributes. Query cap1024 bytes/64 terms; field cap2048 bytes.
The explicit initial ASCII contract rejects other text rather than pretending
Go/Python Unicode semantics agree. Multilingual support remains outstanding.

The optional service research hook supports this map and DirectOrder. Direct
probability ordering differs from the older capped-delta mode, which can flatten
low probabilities. Both preserve backend scores and forecast laws, expose the
research rank delta, and clear experimental packet confidence. Default remains
off. No daemon flag, production enablement or deployment was added.

Tests: all160 feature arrays match Python within1e-14 and probabilities within
1e-12; caller mutation cannot change weights; metadata cannot alter features;
invalid inputs reject. A20-candidate service-boundary test promotes candidate20
while preserving backend score and forecast law. Targeted race tests and vet
passed. This does not establish quality under full-frontier retrieval or LLM use.

## M4 microbenchmarks

Command: go test ./internal/researchsparse -run '^$' -bench . -benchmem -count=3

| Operation | Run1 | Run2 | Run3 | Allocations |
| --- | --- | --- | --- | --- |
| Cached features -> score | 298.8 ns | 297.9 ns | 297.7 ns | 0 |
| Extract+score200 | 1.404629 ms | 1.405186 ms | 1.403418 ms | 7406, about1.103 MB |

Standalone public-pilot input, not worst-case inputs, concurrent p99, database or
agent timings. Extraction dominates cost. Query token/hash reuse is a possible
optimization after full-frontier correctness validation. The seven-direction
goal and production readiness remain unproven.
