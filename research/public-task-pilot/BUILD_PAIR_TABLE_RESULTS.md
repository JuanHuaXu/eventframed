# Atomic build-pair table: no reliable construction win

Status: NOT ADOPTED. Correctness passes, but the discovery timing comparison does
not establish a useful construction improvement. No offered-load run follows.

## Validation

Three full researchindex race repetitions with tags `research_pair_table` and
`research_candidate_only`, using `build-pair-table-overlay-v1/verify-overlay.json`,
passed in17.506s. On every real cache hit, that validation-only overlay recomputes
the original metric and compares exact float32 bits. The active test requires
nonzero calls/hits and checks detachment before search; candidate-only tests also
retain public hydrated-search behavior. The timed overlay omits recomputation.

Both off/on outputs passed the independent partition-screen verifier's hashes,
candidate score/ownership/merge checks and exhaustive-oracle reconstruction.
The eight-partition arms had zero self misses and mean probe recall1 in both
repetitions. Whole-base arms still had one self miss each and recall~0.981-0.982;
this table does not rescue approximate nomination.

## Measured replacement builds

| Layout | Repeat | Off ms | On ms |
| --- | --- | --- | --- |
| Whole base | 0 | 1153.924625 | 1279.098000 |
| Eight partitions | 0 | 86.018709 | 77.477542 |
| Whole base | 1 | 1144.336542 | 1309.402667 |
| Eight partitions | 1 | 75.558250 | 78.669041 |

The partition comparison changes sign between repetitions. These sequential
discovery runs are not a randomized causal estimate. The table removes the
map-cache's dramatic regression but does not demonstrate a robust net speedup.

Across18 small builds:13,178,769 hits and25,969,991 misses (33.6633% hits).
Across4 whole builds:4,517,505 hits and118,080,439 misses (3.6848% hits).
The off-mode sidecar has null Builds because no hook is installed, not because
construction performed no comparisons. Direct mapping evicts collisions; its
hit rate need not equal the earlier first-seen map's hit rate.

## Artifacts and reproduction

- `build-pair-table-off-results.json` and `.pairs.json`
- `build-pair-table-on-results.json` and `.pairs.json`
- `BUILD_PAIR_TABLE_PROTOCOL.md`
- `build-pair-table-screen.mjs` and `build-pair-table-validation.mjs`
- `cmd/research-build-pair-table/main.go`

Run each mode into a fresh output path with:

```sh
go run -modfile=research-candidate-only.mod -overlay research/public-task-pilot/build-pair-table-overlay-v1/overlay.json ./cmd/research-build-pair-table NEW-results.json off
node research/public-task-pilot/check-partition-screen.mjs NEW-results.json
```

Use `on` for the second mode. Existing artifacts and directories are exclusive
and must not be overwritten. No production/module-cache/normal-dependency or
whitepaper change was made.

Next direction: inspect whether bounded incremental/tiered graph maintenance can
avoid rebuilding the same immutable shard at every delta drain. Cache-hit
optimizations alone have not removed the demonstrated construction bottleneck.
Any alternative must preserve global delta capacity, exact version/tombstone
handling, bounded query effort, durable publication and unchanged load gates.
All seven whole research goals remain open.
