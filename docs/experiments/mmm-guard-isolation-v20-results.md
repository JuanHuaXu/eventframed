# Guard-entry isolation v20

Twelve non-race persistent-store cells completed: three trials of writes0/32
crossed with journal prefetch before/inside the guard. Each cell issued64 recalls
from four readers and attempted full50-candidate validation. No ledger work or
labels were involved. Protocol and selected source hashes are embedded in
`mmm-guard-isolation-v20.jsonl`.

## Result

| Writer load | Journal before guard | Journal inside guard |
| --- | --- | --- |
| No concurrent writes | 64/64 accepted in all3 trials | 64/64 accepted in all3 trials |
| 32 future writes | 0/64 accepted in all3 trials | 0/64 accepted in all3 trials |

All384 read-only observations entered, validating19,200 candidates. All384
write-loaded observations were rejected busy; none were stale. No errors or
queue drops occurred. Moving the journal read did not rescue entry. In this
workload the immediate acquisition failure is associated with concurrent owned
writes, not an intrinsically unusable callback or necessary read-only contention.
This does not identify every scheduler phase or prove behavior for other loads.

Read-only acceptance is itself insufficient: accepted-observation p95 age was
461.367-507.490ms, already above the250ms age target used in earlier learning
load studies. This contains repeated50-candidate validation but no model fitting
or ledger I/O. Queued entry alone therefore cannot establish the desired full
learning-path staleness/latency result. Batch validation cost remains a lead.

## Verification

The small factorial accounting fixture passed three race repetitions. The full
12-cell run completed in10.698s and passed accounting/read-only-control checks.
Service vet passed. Independent structured inspection verified all12 unique
cells, embedded source hashes, request/write counts, attempt/drop conservation
and50 validations per accepted observation. Artifact SHA256:
`65f320f310e17fd421dfba577be1307570cd38a4995dd9c5950fd633dffce016`.

Only test-harness options changed for this diagnostic; the v19 entry point keeps
its prior default writer count and prefetch placement. V19 retains its original
embedded source. No publication or lock change was present in the measured v20
run. The subsequently added queued-guard prototype is a separate checkpoint.
