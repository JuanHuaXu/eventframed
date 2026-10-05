# Pruning empty immutable paths

## Confirmed defect and correction

The new `TestLayeredPruneChurn` failed on the unpruned implementation at ID 0:
the deleted-only root retained paths. `TestLayeredPruneSiblingAndMissing` also
failed: deleting absent ID 123 allocated 33 path nodes and changed the root.
These were direct current-root observations, not inference from RSS or GC timing.

The defect was in research-local `layeredReplace`: it unconditionally allocated
every node of the replacement path and retained empty children. No upstream
backend code is involved. The fix returns nil for empty subtrees, reuses branches
whose child pointer does not change, and only allocates changed live ancestors.
Old nodes are never modified. Global-state changes still apply on a no-op tree.
The edit preparation and vector/link semantics are unchanged.

Prior alternatives (delayed GC and old-reader retention) cannot explain an empty
path reachable from the current-only root. A lost sibling or changed historical
snapshot would falsify the patch's ownership invariant; both are tested.

## Verification

- Three targeted race runs passed in 1.369s: 1024 distinct-ID insert/delete
  cycles, eight retained historical roots, sibling preservation across both root
  branches, missing-delete pointer identity, and prior ownership/accounting tests.
- Explicit captured-state replay and storage tests passed under race in 67.815s:
  32 HNSW updates and all 34 historical snapshots still match captured state.
- The final ordinary researchindex race suite passed in 7.928s. Opt-in capture
  tests skip in that ordinary invocation; their explicit result is above.

After each churn deletion, the current-only root is nil and accounts for zero
graph storage. This does not imply immediate GC or zero process RSS. Historical
readers retain their correct original records until released by their owner.

The short captured trace removes only 11 reachable radix nodes (264 structural
bytes) per final current root: 1630 nodes at N800, 12830 at N6400. The main benefit
is preventing accumulation under unique-ID churn, not a large RAM saving in this
16-update trace. Reader/candidate lifetime budgets remain unimplemented.

## Component performance

Same supplied-edit benchmark as the prior report, Apple M4, CPU4, one second per
repeat, no race instrumentation. Fixed 93-record deletion; excludes actual ANN
edit discovery and durability.

| Repeat | Iterations | ns/op | B/op | allocs/op |
| --- | ---: | ---: | ---: | ---: |
| 1 | 8859 | 138052 | 118760 | 3577 |
| 2 | 8468 | 138007 | 118760 | 3577 |
| 3 | 7789 | 138732 | 118760 | 3577 |

This is higher than the earlier unpaired 126448-127887ns/op observations. Do not
attribute the full difference to pruning without interleaved controls. Keep the
correctness fix; do not claim a performance improvement or sustained-load rescue.

## Artifact continuity

`layered_graph_unpruned.go.txt` and `layered_storage_unpruned_test.go.txt` preserve
the earlier source versions and were checked against `layered-storage-hashes.json`.
Those archived paths now supply the old hashes; active sources have changed.
The previous raw JSON results remain unchanged. New outputs are
`layered-go-pruned-results.json` and `layered-storage-pruned-results.json`.

Next: bound live reader/candidate ownership and implement real private ANN update
discovery. All seven whole research goals remain open.
