# Immutable generation lifecycle reference

Implemented internal/researchindex/generation.go as an unwired lifecycle model.
It stores owned vectors in an immutable exact base plus a bounded delta; it is
not yet an ANN index or a durable backend. It is evidence about publication
semantics, not throughput or an implemented replacement for libravdb.

## Invariants exercised

- Prepare copies vectors and materializes the next generation while holding
  writer exclusion. Readers continue seeing the previous generation.
- Commit publishes the prebuilt pointer and releases exclusion; the isolated
  publication path measures zero allocations. Abort publishes nothing.
- Successful preparation must be settled exactly once. Cancellation after a
  hypothetical durable commit cannot prevent publication. The model does not
  itself establish that the hypothetical commit happened.
- A captured old view stays unchanged, including across deletes and compaction.
- One compaction may build at a time. Publication carries forward delta entries
  newer than its captured revision, including tombstones and overwritten IDs.
- Compaction leaves semantic revision unchanged and reclaims incorporated delta
  entries. At the declared delta cap, new IDs are rejected without mutation.
- Concurrent writer, compactor and reader loops preserve view consistency and
  the final committed revision. Input and output vector mutations do not alias
  published records.

The lifecycle-to-ranking test runs600 deterministic update/delete/abort steps
and overlapping compactions, compares every top10 with a full committed-history
oracle, and rechecks retained old views. The scorer is exact dot product, not
semantic embedding evaluation or evidence that HNSW finds the same prefix.
The full researchindex package passes five race-enabled repetitions with
`go test -race ./internal/researchindex -count=5 -timeout=120s`. This includes
the earlier2000-case merge oracle and its missing-prefix negative control.

## Still missing

There is no WAL integration, crash recovery, real HNSW base builder, off-heap
reclamation, compaction scheduler, corpus-scale benchmark, or service replay.
Compaction currently materializes an exact map with corpus-scale work; it does
not make this work disappear. Go GC keeps old views alive, but callers retaining
unbounded views can retain unbounded memory. Prepared writers can block later
writers until settled; handling abandoned owners needs a larger lifecycle design.

Next wire a durable research transaction to prepare/commit/abort with explicit
unknown-outcome handling, then replace the reference base with a real immutable
search index. An uncertain storage error must not be treated as an abort merely
because this in-memory model makes abort easy. All seven whole goals remain open.
