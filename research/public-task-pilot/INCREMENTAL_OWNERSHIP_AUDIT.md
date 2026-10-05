# Incremental HNSW ownership audit

Status: direct prepared-HNSW adapter is NOT ready. Existing dynamic operations
are useful algorithmic components but do not supply immutable graph versions.

## Authoritative code findings

Pinned local copy: `candidate-libravdb-v1.6.13`.

- `internal/index/interfaces.go:54` requires PreparedMutation.Commit to avoid
  allocation and search. At:65, DeltaIndex requires PrepareMutations. The HNSW
  wrapper at:174/:218 delegates directly to mutable Insert/Delete and has no
  PrepareMutations; the flat wrapper at:571 does implement it.
- `libravdb/tx.go:1460` selects that optional DeltaIndex interface. HNSW does not
  enter this preparation path. Adding a method that calls Insert during Commit
  would violate the interface's explicit contract.
- `internal/index/hnsw/hnsw.go:133` owns shared nodes, ID maps, global entry-point
  state, vector stores, allocator lists and reclamation state. Insert at:459
  allocates metadata, traverses/edits the graph, handles in-flight publication,
  and can retire failed insertion storage.
- `internal/index/hnsw/delete.go:38` changes live node/index state. Incoming-link
  repair at:110 edits link slots and counts in place using atomic stores. Copying
  only the outer Index struct would still share this mutable state and allocators.
- `internal/index/hnsw/reclamation.go:108` registers reader epochs; :180 checks
  when retired allocations can be reclaimed. This protects allocation lifetime,
  not a queryable historical root containing old vectors, links and ID mappings.

The audit script asserts selected signatures and stores source hashes, but those
assertions are sanity checks, not a formal proof of all paths.

## Upstream and tests

Fresh public GitHub API snapshot in `incremental-ownership-audit.json` reports
master `18515d4ce0620b24fd2512b0f9023b8b1f4d8355`, unchanged from the prior audit.
Seven pull requests and10 issue entries (including PRs), no further pagination.
No newly listed prepared-HNSW implementation was identified. Titles/listings are
not a semantic review of every patch; the pinned source remains the code basis.

Existing backend tests were executed, unmodified, in the isolated copied module:

```sh
go test -race ./internal/index/hnsw -run 'TestDeleteRepairsAsymmetricIncomingLinksBeforePrune|TestReclamationConcurrentSearchDeleteReinsert' -count=3
```

PASS,1.817s. These tests support current deletion/reader reclamation behavior,
not transaction rollback or historical snapshot isolation. No module/source edits.

## Next experiment

Prototype a bounded persistent graph-state kernel outside the backend: immutable
node/vector/link records, a versioned root, prepared replacement of touched nodes,
and atomic root publication. Test old-root traversal after insertion, deletion,
abort and newer publication; ensure no post-commit allocation/search. Charge
copied adjacency/vector bytes and retained roots explicitly. This kernel would
establish ownership only, not HNSW quality or update speed.

Then instrument actual HNSW insertion/deletion to measure the touched-state set,
including backlinks and entry-point changes, before deciding whether that kernel
can host its update algorithm efficiently. Fixed degree does not itself bound all
incoming neighbors or repair work. A full graph copy is also not a scalable fix.
Do not install a direct in-place adapter merely because current concurrency tests
pass. All seven whole research goals remain open.
