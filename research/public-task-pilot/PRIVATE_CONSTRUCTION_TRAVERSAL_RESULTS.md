# Private construction traversal

Status: component implemented; backend construction equivalence is unverified.
This does not complete private insertion or rescue sustained-load serving.

`SearchConstructionLayer` now shares traversal with `SearchLayered`, but takes
an explicit entry ordinal and level. It does not descend, apply a public-query
ef floor, or include concurrently in-flight nodes. Candidate distances use the
existing scalar cosine metric and are converted to float32 for private selection.
The evaluation budget returns no partial candidate set on exhaustion. This is
not a total graph-read or retained-byte bound.

Verification:

- Three race-enabled repetitions of construction level isolation and existing
  retrieval budget tests passed (1.330s). Cases include literal ef=1/2,
  non-global entry, unavailable level/entry, empty graph, cancellation and
  exhausted evaluation budget.
- Ordinary researchindex suite passed (4.981s); optional captures are skipped
  there, so the backend comparison was run separately.
- Explicit backend-query capture replay passed (1.911s). All four graphs
  retained 320/320 matching top-ten IDs across 32 queries at ef=400. Exact-oracle
  hits remain 320/320 for N=800 and 319/320 for N=6400, initial and final states.
  Maximum evaluations remain 920/911 and 5990/5957 respectively. These are
  previously exercised development queries, not fresh confirmation.

Next: compare construction candidate/selection outputs against the pinned
backend on identical snapshots, including tie/order behavior. Public query
agreement does not establish construction agreement, especially when selection
returns a short candidate list without sorting. Then compose insertion and its
overflow repair before durable integration and sustained-load retesting.
