# Construction backend comparison

Scope correction: this initial comparison imposed a level-zero selection cap
of 8; the default backend selector uses 9 for M=4. The subsequent
[link decision study](PRIVATE_LINK_RESULTS.md) records a fresh capture with that
actual limit and another zero-mismatch comparison. Do not read the initial cap
as the default insertion configuration.

The corrected serial 40-node, 768-dimensional backend capture exercises 90
cases: three query vectors, ef=1/3/8/16/32, and all six levels of the entry node.
M=4, level-zero maxM=8, seed=65537. Vectors are regenerated and hash-checked;
pair distances for selection come from the backend capture, not an independent
metric implementation. No production service or backend dependency was changed.

## Findings and repairs

1. The first harness retained a selected slice backed by pooled scratch. Later
   calls overwrote it. `construction-capture.json` and `construction-comparison.json`
   are INVALID for selected-neighbor conclusions, retained as failed artifacts.
   The harness now uses caller-owned scratch and copies before release.
2. Corrected capture `construction-capture-v2.json` exposed one genuine traversal
   mismatch: level 1, ef=3, query 1 missed ordinal 8. Backend traversal explores
   backlinks as well as outgoing links; the private traversal omitted backlinks.
   `construction-comparison-v2.json` records that failed comparison.
3. Shared traversal now explores both adjacency lists through one seen set.
   Backlink-only regression checks cover both construction and query paths.

## Verification

`construction-comparison-v3.json`: all 90 candidate sets, selected sets and
selected ordinal orders match. Race-enabled comparison plus level/budget tests
passed in 1.324s. This is finite fixture evidence, not universal equivalence;
concurrency, SIMD near ties, candidate modes and larger construction fixtures
remain outside this comparison.

Public-query capture replay retained all 320/320 matching top-ten IDs in each
of four graphs (32 queries per graph). Exact-oracle hits remain 320/320 for
N=800 and 319/320 for N=6400. Maximum metric evaluations increased from
920/911/5990/5957 to 948/921/6127/6069. This additional work is required by the
backend's navigable adjacency, not a demonstrated latency improvement.

Private insertion composition, connection overflow handling, durable integration
and the original sustained-load gate remain unfinished. No whole research goal
is completed by this comparison.
