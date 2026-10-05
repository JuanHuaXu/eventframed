# Background process-exit recovery v3

TestDurableBackgroundRecoveryParity connects Background.Record, the actual SQLite
ledger and original-record replay. An uninterrupted fixture provides the control.
Separate child processes persist64 admissions and64 feedback records before
terminating without Close. The queued case holds fitting so no label can finish;
the applied case waits for64 completed updates before exiting. Each child must
return the expected exit code24; a timeout or test failure is not counted as a
successful boundary test.

Both cases passed three race-test repetitions; vet passed. Reopening and replaying
17-record pages yields64 labels, zero pending and128 log entries. All512 final
feature-state forecasts match the uninterrupted background control exactly.
Retrying every persisted record returns the same sequence as a duplicate and
does not cause a second application in this harness.

This combines previously separate ledger and asynchronous replay tests, but
does not supply a live durable-consumer API. It rebuilds from the complete log
into a fresh Adapter, so no checkpoint/applied-position transaction is involved.
The fixture deliberately keeps original predictions cold while feedback queues;
mixed partial progress and warm-model admission still need recovery coverage.
Service dependency authority, tenant/contract validation at the consumer boundary,
discard/pending recovery, log retention, replay cost and I/O failure handling
remain open. Process exit is not power loss or interruption inside COMMIT.

No real-task quality improvement, deployment or production readiness is claimed.
