# Persistent replacement-entry summary

`entry_summary.go` implements a persistent32-bit radix selection index. Each
subtree stores its highest-level candidate, breaking ties by lowest ordinal.
Updating a candidate's level, inserting it, or deleting it copies at most33 nodes;
reading the best replacement is O(1). Empty subtrees are pruned.

This summary supplies a replacement when the current entry point is deleted.
It is NOT a rule to replace the graph's entry point after every insertion:
the backend keeps an existing tied entry point. The caller must maintain and
publish summary and graph roots coherently. That integration is not implemented.

## Tests

- 3000 deterministic randomized insert/level-change/delete operations compare
  every summary answer with an independent full-map scan; retained historical
  summaries, high-bit ordinals, invalid levels, and empty pruning are covered.
- The serial controlled32-operation capture checks summary results before/after
  each operation against independent scans. Both forced entry-point deletions
  also match the actual backend's packed replacement ordinal.
- Corrected explicit capture plus oracle race run passed in19.108s.

The first three capture runs failed because the harness read the low32 bits of
packed global state as the ordinal. `global_state.go` defines the ordinal in the
high32 bits and level+1 in the low32. Correcting the assertion made the capture
test pass; the selection algorithm did not change. This diagnostic error is not
a backend or summary failure.

## Interpretation

This provides a tested way to replace the observed808/6408-slot fallback scan
with maintained selection metadata. It adds update/storage overhead; neither
end-to-end latency improvement nor durable integration is established. It must
be wired into real private update preparation, not maintained from a whole-graph
diff in serving. Corpus-wide initialization in the replay is fixture setup only.

The other deletion repair work, insertion discovery, byte budgets and durable
publication remain unresolved. All seven whole research goals remain open.
