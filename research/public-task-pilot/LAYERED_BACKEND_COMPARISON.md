# Identical-state backend traversal comparison

Source inspection found a material configuration difference: the backend's public
search uses `max(EfSearch, k, 2*EfConstruction, efOverride)`. Here EfConstruction200
makes effective breadth400 even when callers request100 or200. The private
prototype previously used the requested breadth directly. Earlier failures remain
valid for their actual budgets, but were not equal-effective-budget comparisons.

The new backend harness runs the same32 absent-ID queries on serially constructed
initial/final graphs. Every graph hash matches the earlier capture used by private
search. The private comparison uses the observed backend effective ef400, not a
post-hoc search over breadths until the target passes.

| N | Snapshot | Backend exact hits | Private exact hits | Shared result IDs | Max private evaluations |
| --- | --- | ---: | ---: | ---: | ---: |
| 800 | initial | 320/320 | 320/320 | 320/320 | 920 |
| 800 | final | 320/320 | 320/320 | 320/320 | 911 |
| 6400 | initial | 319/320 | 319/320 | 320/320 | 5990 |
| 6400 | final | 319/320 | 319/320 | 320/320 | 5957 |

Matching counts establish the same top10 SETS in every query, not necessarily
identical ordering or floating-point scores. Both are still approximate and miss
one expected neighbor per6400 snapshot. Backend test passed in5.552s and comparison
in1.972s; neither mixed setup/oracle elapsed time is a query latency benchmark.

## Interpretation

On these now-observed development queries, private traversal is compatible with
the backend's result sets at matched effective breadth. This does not independently
validate generalized retrieval or establish that the algorithms are equivalent.
The public API's search quality floor is not automatically the correct rule for
insertion-time construction; that path has a separate EfConstruction contract.

Quality improved with a large work cost: nearly6000 metric evaluations for6400
records. Frontier size is not evaluated-corpus size, and these results do not
support cheap corpus-independent ANN search. Before serving integration, compare
latency and growth under explicitly matched budgets. Preserve ef100/200 failures;
do not relabel them successful or reuse these queries as untouched confirmation.

Artifacts: `backend-heldout/results.json`, `backend-heldout/comparison.json`, and
the backend/private harnesses. All seven whole goals remain open; production and
standard dependencies are unchanged.
