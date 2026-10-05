# Prepared durable wrapper load v41

Frozen before execution. Twelve rotated cells: three trials of off, non-durable
group4, post-guard verification/discard with original batch SQL, and the same
path with transaction-scoped prepared batch SQL. The explicit PreparedWrites
flag distinguishes the two durable cells; Mode remains group4postverify for both.

Keep 192 recalls/four readers, 96 future writes at 2ms, 50 candidates, ready
groups of at most four, queue 64, 20ms entry deadline and fresh persistent stores.
Prepared writes enter only through an opt-in owner constructor; default remains
unchanged. Replay and single-record operations do not change. Service validation,
actual owned-worker admission and original-return comparison remain guarded.
Full ledger readback and typed discard remain post-guard; count completion/age
only after both finish. FULL commits, pending cap and fail-closed errors remain.

Run warm 200-record parity, original/terminal reopen retries, uncertain-commit
recovery and small-load race tests first; then full ledger/learner/service race
and vet. Preserve exclusive JSONL with selected source hashes, raw timings,
PreparedWrites flag and complete operation counts. No labels/fits in load cells.
Warm parity must match the original forecast-producing history: wait for each
training update and compare persisted training originals across owners. Merely
matching labels and waiting for the final count leaves scheduling confounded;
an explicit delayed-fit control must demonstrate that distinction.

Screens stay 250ms accepted-age p95 and read-p99/off <= 1.10 per paired trial.
Report writer tails, queue drops and expiries separately. Passing isolated v40
does not imply passing this test. A finite pass here still does not establish
warm loaded learning, feedback/history authority, realistic answer quality or
all-direction completion. No production configuration change or publication.
