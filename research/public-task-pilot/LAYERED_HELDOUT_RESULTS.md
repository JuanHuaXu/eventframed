# Held-out synthetic traversal result

Both ef configurations FAIL the predeclared95% recall target at N6400. All N800
cases are exact. The test command completed normally (2.055s package elapsed),
which means the experiment ran, not that the quality criterion passed.

| Corpus | Snapshot | ef100 hits | ef200 hits |
| --- | --- | ---: | ---: |
| 800 | initial | 320/320 | 320/320 |
| 800 | after mutations | 320/320 | 320/320 |
| 6400 | initial | 230/320 (71.875%) | 292/320 (91.25%) |
| 6400 | after mutations | 230/320 (71.875%) | 292/320 (91.25%) |

These use the same32 predeclared absent-ID random vectors for each graph and
breadth, k10, evaluation limit20000. All queries completed within the evaluation
limit; loss is approximate-search recall, not an unreported budget error. Raw
per-query hits/evaluation counts are in `layered-heldout-results.json`.

This materially weakens the earlier317/320 self-vector smoke result as evidence
of readiness. Self queries can find their own neighborhoods while new queries
still miss relevant nearest neighbors. The vectors are synthetic numeric data;
no semantic/domain-transfer claim is made. Queries are now observed development
data, not a future untouched confirmation set.

Do not immediately increase breadth until this fixed set passes. Next compare
the real backend on identical graph states and queries, isolating private-traversal
behavior from limitations of the underlying graph. Insertion integration must not
silently treat this reader as validated. All seven whole goals remain open.
