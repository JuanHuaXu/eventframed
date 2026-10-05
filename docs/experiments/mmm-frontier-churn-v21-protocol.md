# V19 frontier churn v21: consumed-fixture diagnostic

Date: 2026-10-02. This diagnostic uses the already-consumed v18/v19
loaded Recall fixture. It is not an independent confirmation cohort,
a posterior-reuse authorization, or a new scored-quality test.

Reconstruct the exact cosine ordering of the fixture's two angle-zero
seed rows, 198 eligible rows with angles `.005*(i+1)`, 16 future-only
rows, and 128 visible appends with angles `.00213*(i+1)`. Exclude the
future-only rows at the fixed request as-of. Before each visible append,
freeze exact top-10, top-50 and top-150 sets; compare with the sets
after that append. Independently check that the generated 256-dimensional
vectors have cosine similarity `cos(angle)` to the query within float32
tolerance. Report set changes, first non-changing append, final cutoffs,
and compare the widths without inferring actual packet packing or Brier.

The question is whether v20's exact top-150 frontier-equality guard can
often accept posterior reuse under the v19 write stream. A narrow top-10
set may stabilize earlier, but that alone cannot certify reuse: the
selection-conditioned likelihood and candidate activation probability
may depend on the larger nomination frontier. The v18 feedback fixture
also labels only 16 selected outcomes, all `Useful=true`; it cannot
estimate proper-score improvement, false-positive cost, or learned
calibration for a relaxed guard. Preserve that limitation.
