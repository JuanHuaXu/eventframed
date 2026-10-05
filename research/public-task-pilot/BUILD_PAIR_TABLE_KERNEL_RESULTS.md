# Atomic pair table: kernel evidence only

Status: candidate for integration, NOT a construction or serving rescue.

The preceding mutex/map integration remains rejected: its 39.30% hit rate
accompanied roughly fivefold replacement-build slowdown. Lock contention,
heap-escaping per-comparison callbacks, and cache locality are plausible costs;
their individual contributions are not established by the end-to-end result.

`internal/researchindex/build_pair_table.go` uses one atomic 64-bit word per
slot to publish a full directed 16-bit pair tag and exact float32 bits together.
Collisions evict, never alias. Unsupported IDs bypass caching. Capacity is a
power of two capped at 65,536 slots (512 KiB slot storage plus counters).
It must remain local to one immutable-vector build. This encoding is not a
claim of caching coverage for arbitrarily large node IDs.

## Correctness

Command:

```sh
go test -race ./internal/researchindex -run TestBuildPairTable -count=3
```

PASS, 1.357s. Covers empty keys, directed keys, collisions, unsupported IDs,
signed zero and NaN payload preservation, concurrent eviction without false
hits, counter accounting, and zero per-operation allocations. No HNSW hook is
activated by these tests; lifetime and real-distance integration remain next.

The full ordinary researchindex suite also passed three race-enabled runs:
`go test -race ./internal/researchindex -count=3` (16.868s). Build-tagged overlay
integration tests are not included in that command.

## Bookkeeping experiment

```sh
go test ./internal/researchindex -run '^$' -bench BenchmarkBuildPairBookkeeping -benchmem -benchtime=200ms -count=3 -cpu=4
```

Apple M4, darwin/arm64. PASS, 8.246s. Values below are all three ns/op results;
every row reported 0 B/op and 0 allocs/op.

| Pattern | Mutex/map | Atomic table |
| --- | --- | --- |
| Serial miss | 6.233, 6.249, 6.169 | 2.727, 2.717, 2.719 |
| Serial hit | 4.088, 4.093, 4.087 | 1.772, 1.819, 1.773 |
| Parallel miss | 126.3, 122.5, 120.9 | 13.30, 12.94, 13.42 |
| Parallel hit | 58.08, 62.08, 104.3 | 13.17, 13.07, 13.86 |

This intentionally reuses one pair, maximizing shared-counter/lock contention.
Misses use a full non-evicting map and an empty table with no insertion. The map
executes a constant callback, while the table only looks up. Thus these are
bookkeeping diagnostics, not equivalent complete cache-miss pipelines. In
particular, the benchmark callback does not recreate the escaping integration
closure, and zero allocations here do not disprove that integration cost.

Next: activate Lookup/Store in a separate fresh-build overlay, verify original
distance bits and callback lifetime, then run matched cache-off/on construction
and exact-oracle gates. Reject if net construction cost or correctness regresses.
No production source, normal dependency, or existing frozen result was changed.
All seven whole research goals remain open.
