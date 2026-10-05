# Rolling-LSN reader v3: frozen sortable-key visible-write screen

The [v2 timing screen](mmm-rolling-lsn-reader-v2-results.md) passed on
whole-second data but failed a separate fractional as-of falsifier because
the existing variable-width `available_at` metadata is not lexically
sortable. A [raw-store probe](mmm-lsn-fixed-key-v1-results.md) passed with
a new fixed-width UTC key. This v3 screen tests that key through the
research-only versioned batch writer and under the same loaded reader/write
shape. Production collection, writer, search and daemon remain untouched.

The research writer must derive `available_at_sort` atomically from the
EventFrame's authoritative `AvailableAt`, using exact nine-digit UTC
fractions. An exact retry over a row missing or disagreeing with that key
must fail closed. Mixed/partial migration is not acceptable. The private
collection declares both the ordinary and sort-key fields as strings.
Use the captured-snapshot/exact-LSN linearization rule from v2; no
post-read current-version retry.

For each of three fresh rotated pairs, seed 200 records available at
`00:00:00.100Z` and 16 future-only records at `00:00:00.125Z`. Fix
read as-of at `00:00:00.120Z`, so the old variable-width SQL predicate
would omit past rows and admit future ones. Offer 256 visible records at
the read as-of in 16 ordered batches of 16 at nominal 16 ms intervals;
offer 192 k=50 reads at nominal 4 ms to eight workers. Use the same
base/future/visible vectors and ranking checks as v2. The candidate SQL
predicate is `available_at_sort <= $available_by_sort`; control remains
ordinary `Store.Search`. Every read must return 50 eligible distinct IDs,
the candidate may not show a row newer than its captured version, and
no future ID may appear. Require all 256 acknowledgments and per-version
motion, a final latest-batch read, zero active leases, and an after-timing
check that every stored event has the correct derived sort key. Reopen and
verify the sort key on a fresh query. Freshness counts only reads offered
after each batch acknowledgment.

Frozen component gate: no errors or omitted requests; pooled candidate
read-call p99 <=1.10x control and <=100 ms; summed writer completion
<=1.25x control; candidate post-ack visible-read lag p99 <=25 ms. Report
pair and pooled values, actual offer gaps, writer acknowledgment p99,
retained bytes and worst lag. A pass is a research backend component,
not Goal 6 completion; existing collection migration, SQLite authority,
learner publication, full Recall latency and power-loss recovery remain.
