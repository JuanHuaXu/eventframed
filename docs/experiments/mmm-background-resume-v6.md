# Background continuation after replay v6

ResumeBackground rebuilds an exclusively owned quiescent ledger, then transfers
the model, unresolved original records, next ID and feedback time into a worker.
Pending records move to background ownership; the fitting adapter starts with
an empty pending map so discard cannot leave a hidden second copy. Completed
count resumes at the number of replayed labels. No new goroutine starts on a
replay error or invalid capacity.

TestResumePendingAndContinue restores64 labels and16 unresolved predictions,
checks each original record, discards one, processes15 delayed labels, and checks
all512 forecasts against the uninterrupted control. It verifies zero hidden
pending records, rejection of discarded feedback, next ID81 and wrong-epoch
rejection. Three full researchmemory/researchledger race repetitions and vet pass.

This is a continuation primitive, not a durable-consumer service. New predictions,
feedback and discards are not automatically written; a wrapper must durably
record their ordering before acknowledgment. In particular, the test discard
does not survive a second replay until a durable discard operation is defined.
Service event identity, dependency authority, retention/checkpoints and loaded
durable-write latency remain open. Old performance artifacts retain their own
embedded worker versions and are not retroactively revalidated by these tests.
