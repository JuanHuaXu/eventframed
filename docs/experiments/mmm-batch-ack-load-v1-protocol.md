# Batch acknowledgement load v1: frozen backend screen

This is a Goal 6 research-only **backend contention** screen. It does not
replace the unchanged full-Recall/learner 4 ms gate. Use the real-store
batch-intent prototype under one process and exclusive temporary stores.
Do not modify production service, publication, lineage or daemon wiring.

For each arm, first seed 200 past-available same-tenant EventFrames through
authority-protected batches. Then concurrently offer 256 future-only writes
at 1 ms nominal spacing and 192 backend vector searches at 4 ms nominal
spacing with eight reader workers. A bounded one-writer handoff accepts one
write per job; the control commits jobs singly. The candidate may commit up
to 16 compatible jobs per batch after a maximum 4 ms coalescing wait.
Every job receives one result only after durable batch finalization. Record
offer-to-ack, writer call, read call, actual offer gaps, accepted/duplicate
counts, total write completion and batch-size distribution. Assert every
reader result was available by the requested as-of time and the final
snapshot is exactly +256 with all per-version motion present. No dropped or
false acknowledgements. Run control/candidate in rotated order on at least
three fresh matched pairs with the same fixture IDs and search vectors.

Frozen screening criteria: candidate median full-writer completion must be
lower than control; candidate pooled offer-to-ack p99 must not exceed 1.25x
control; candidate pooled read-call p99 must not exceed 1.10x control. Report
all per-pair values even if pooled gates pass. These are diagnostic thresholds,
not evidence of full Goal 6 completion. A candidate failing any gate is not
advanced to service integration without a new frozen design.

Use deterministic corpus/query generation and no future outcome labels.
Reader queries are read-only and their result identities cannot depend on
future-only writes. The 4 ms read offer rate can be missed under overload;
record actual gaps and do not present the run as a fixed-rate service test if
that occurs. The synthetic hash-vector corpus, raw backend search instead of
full `Service.Recall`, absent Bayesian learner and SQLite journal, and no
cross-process power-loss test limit any positive result.
