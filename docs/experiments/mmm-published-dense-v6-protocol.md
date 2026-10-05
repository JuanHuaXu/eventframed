# Dense 256D published-LSN top-k v6: frozen transfer protocol

Date: 2026-10-02. Research-only, opt-in. No production writer or reader
change. The v5 4D top-k result is not reused as high-dimensional evidence.

## Fixture

Create a private 256D cosine collection with declared EventFrame payload
and sortable availability columns, using the existing journaled publication
gate but a separate 256D setup. The deterministic dense unit query has
coordinate `sin(0.17(j+1)) + 0.3 cos(0.071(j+1))` for `j=0..255`, normalized.
For row i, derive a deterministic dense noise direction from
`sin(0.113(i+1)(j+1)) + cos(0.047(i+3)(j+1))`, project it orthogonal to
the query, and normalize. The row vector is twice the unit mixture
`cos(theta_i) q + sin(theta_i) n_i`; a test-only boundary normalizes the
stored vector before committing it. Fixed eligible angles are
`theta_i=0.005(i+1)` for `i=0..197`. Two genesis eligible rows have
angle zero. A genesis future sentinel has angle zero; 16 more future
rows use `theta_i=0.001(i+1)`. Eligible availability is 100 ms and
the as-of is 120 ms after 2026-10-02T00:00:00Z; future availability is
125 ms. The raw query is `2q` and is normalized before SQL.

## Comparison

After journal-certified publication, compare SQL cosine top-k with full
EventFrame projection against ordinary `Store.Search` at k=10, 50, 200
on the same committed corpus. Require exactly k unique as-of-eligible
rows, no future row, exact set equality (tie order immaterial), each
similarity within `1e-5`, and nonincreasing SQL similarity. Warm four
times, then measure 32 alternating quiet calls per arm/cap and report
median and nearest-rank p99. Freeze the fixture and gates before the run;
do not tune HNSW, angles, seeds, k, or tolerance in response to failures.

Run focused normal/race tests, ordinary package tests, and vet. A failure
is preserved as negative transfer evidence. The quiet timings exclude
ingestion, publication, journal, full Recall, network, and offered-load
queueing. Even a pass is not a loaded Goal 6 or large-corpus result.
