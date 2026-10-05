# Retained layered-state storage

## Evidence

`TestLayeredStorageAccounting` and `TestLayeredCapturedStorage` passed under
`go test -race` in 22.369 seconds using the corrected captured HNSW updates.
The test-only walker deduplicates immutable trie nodes, records, vector backing
arrays, adjacency backing arrays, heuristic arrays, and layer-header backing
arrays by pointer identity. It counts slice capacity, not only slice length.

`layered-storage-results.json` holds the counts. There are 16 updates per corpus.
The recent-reader scenario retains the final root and the preceding eight roots;
the old-reader scenario retains the final root and the first eight roots. These
are selected retention patterns, not an exhaustive worst case for eight readers.

| Corpus | Current only, bytes | Current + 8 recent | Current + 8 oldest | All 17 roots |
| --- | ---: | ---: | ---: | ---: |
| 800 | 2,961,128 | 3,220,344 | 3,483,536 | 3,581,920 |
| 6400 | 23,699,544 | 24,085,608 | 24,383,136 | 24,493,512 |

At 6400 records these two eight-reader scenarios add about 1.63% and 2.88%
relative to current-only storage. At 800 they add about 8.75% and 17.64%.
Most vectors remain shared. This supports the ownership design in this trace,
but it does not bound memory as update count or reader age increases.

## What the numbers omit

These are structurally reachable byte counts using Go type sizes and backing
capacities on this host, not RSS, heap-in-use, or allocator measurements. They
exclude allocator rounding, GC bookkeeping, temporary/unreachable preparation
allocations, diagnostic visited maps, and underlying ID string source buffers.
String headers are included in record sizes. A production budget must account
for these omitted owners and for candidate preparations before admission.

The walker scans full retained trees and allocates visited sets. It is an offline
diagnostic, not a proposed per-request admission routine. There is no reader-lease
manager or enforceable live-byte bound in this model yet.

## Confirmed remaining defect and next step

The single-record delete probe leaves 33 empty radix nodes reachable from the
current root with zero records/vector values. Thus reader release alone cannot
reclaim every historical allocation: sequential unique-ID churn retains paths.
The next implementation needs path pruning, tested against historical snapshots,
and bounded reader/candidate ownership. This is a research-model limitation, not
a newly established defect in production libravdb.

Do not infer that the previous sustained-load failures are fixed. Real ANN edit
discovery, durable publication, and long-load validation remain required. All
seven whole research goals remain open.
