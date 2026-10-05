# Known-non-entry queue refinement

Use the same32-arm clean batch queue protocol, deadlines and durability. Change
only the classification of admission failure using the synchronous local
RunAdmitted adapter: if Gate.Write never invoked the callback, the queue may
continue with later jobs. A plain deadline error from an entered callback still
stops the queue, as does every other unclassified commit error. No error-text
matching. Underlying HNSW and storage modules remain unchanged.

Use phase-aware-overlay-v1 and the generated research-task-phase-durable runner.
Output phase-queue-results.json and private retained stores, never overwrite.
Compare descriptively with batch-queue-clean-results.json. Require the same
no-errors/no-late/no-stale/warmup/reopen gates. Refute complete rescue if any
high-rate arm fails. No workload reduction or completion before durable settlement.

Before measurement the overlay queue/admission tests pass ten race repetitions:
never-entered cancellation may continue; entered failures still poison; errors
that merely contain the non-entry wording confer no authority. Existing bounded
queue, settlement and shutdown tests also run. No concurrent agent-started tests
or builds during the measured run. This is not a backend rollback guarantee.
