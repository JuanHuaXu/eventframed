# Merge schedule accounting and dynamic-index leads

## Executed schedule model

`model-run-schedule.mjs` simulates128 append batches of32 entries, with complete
event logs in `run-schedule-model-results.json`. It contains executable assertions
for record conservation, ideal rewrite work and bounded-slot termination. No
latency model or timing fitted from the failed experiment is used.

Repeated growing-tail merging rewrites264,160 entries. Ideal binary equal-size
tiering rewrites28,672, about9.2x less, but eventually builds one4096-entry run.
With eight fixed base graphs and the existing15-graph-plus-delta limit, only
seven tail slots remain. Greedy equal-size tiering stops before batch128: the
127 prior batches occupy tails32,64,128,256,512,1024,2048 (4064 records). There is
no physical slot for the next flush before merges begin.

Allowing an eighth tail would exceed that current contract. Even that hypothetical
case either builds4096 entries or stops at a specified merge-size cap. With512
as a cap, this tested schedule stops after accepting1024 records; with1024 it stops
after2048. These are failures of the modeled greedy policy, not a proof that every
possible tiering or partition policy fails. Future preemptive unequal merges or
base absorption need their own resource and construction-cost accounting.

## Primary research, inspected mechanisms

Singh et al., *FreshDiskANN* (2021),
[author-hosted paper](https://suhasjs.github.io/files/freshdiskann-arxiv.pdf),
sections5.1-5.6: a mutable in-memory temporary index accepts inserts while older
temporary indexes become read-only. Queries combine temporary and long-term
indexes and filter deletions. StreamingMerge updates the long-term graph rather
than rebuilding it from scratch; redo logging supports recovery of mutable state.
Its machines and buffer sizes differ substantially from our workload. We cannot
import its throughput or recall numbers into EventFrame's guarantees.

Xu et al., *In-Place Updates of a Graph Index for Streaming Approximate Nearest
Neighbor Search* (2025),
[primary HTML](https://arxiv.org/html/2502.13826v1), section3.3, Algorithm5: search
near a deleted point, identify visited nodes pointing to it, repair their outgoing
links using candidate neighbors, add selected replacement links for its outgoing
neighbors, then remove the point. The paper reports streaming recall experiments
and targets avoiding batch deletion consolidation. This is not an implementation
of libravdb's transactional interface, nor proof of historical snapshot isolation.

## Implication for the next prototype

The stronger lead is bounded incremental graph editing or merge work that can
continue ingesting, rather than another unbounded from-scratch tail build. Before
editing the backend, inspect its prepared-mutation and graph ownership APIs:

- What exact vectors and neighbor lists change during one insertion/deletion?
- Can a private prepared graph version share untouched state safely?
- Can all fallible edits finish before the authoritative commit, leaving an
  allocation-free publication operation afterwards?
- Can old readers retain a matching graph/data revision without seeing future
  candidates or a partially edited traversal structure?
- How are uncertain durability, graph reclamation and concurrent resource bounds
  reconciled without expanding the pending buffer silently?

Filtering future IDs at output alone does not establish as-of traversal behavior:
future edges may still influence which older candidates are found. Simply turning
on in-place mutation would therefore bypass an existing contract. No such change
was made. Need an API/ownership audit and a bounded reference experiment before
claiming these publications rescue the implementation. The separate admission
failure cluster still needs tracing. All seven research goals remain open.
