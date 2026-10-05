# Ledger abrupt process exit v2

TestAbruptExitCommitBoundary launches isolated child test processes against
temporary SQLite paths. Children terminate with os.Exit(23), bypassing Close
and deferred rollback. Four cases cover admission/feedback with an uncommitted
insert or after successful Append/COMMIT. For feedback, admission is committed
first. Parent verifies the exact exit code, reopens and retries the record.

All four cases passed in three race-test repetitions; package vet passed.
Committed records survive and retry returns the same sequence as a duplicate.
Uncommitted records are absent and retry creates one record in expected order.
Every recovered database passes PRAGMA integrity_check. No production processes,
data or configuration are used.

The uncommitted case uses test-only direct SQL to place execution between insert
and commit; it does not inject failure inside Append. These tests establish
selected process-exit boundaries, not power-loss safety, partial-sector writes,
commit syscall interruption, corrupt files or error acknowledgment behavior.
No learner runs in the child. Durable background acknowledgment/application and
restored-prediction comparison across abrupt exit remain required.
