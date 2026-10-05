# Durable wrapper uncertain writes v9

A private Get/Append/Close interface lets tests wrap the real ledger without
replacing production constructors. The wrapper either returns an injected error
before writing or commits through the actual SQLite ledger and then returns the
same error. Four cases cover admission/feedback before/after commit.

All cases pass three targeted race repetitions, three full researchmemory and
researchledger race repetitions, and vet. Each uncertain append stops the live
wrapper: further admission and feedback reject. Close/reopen resolves actual
durable state. A committed admission restores one pending original record;
an uncommitted admission retries fresh. Committed feedback replays one learned
label and retries as a duplicate; uncommitted feedback leaves one pending record
and the retry learns exactly once. Failure counts remain zero.

This tests the caller's response to commit ambiguity with real persisted data.
It does not inject physical disk failure, torn writes, SQLite commit internals,
multi-process contention or corruption. Reopening scans the full trusted log;
there is no checkpoint or bounded restart-cost claim. Durable write performance,
discard, service dependency/event binding and broader predictive-quality work
remain incomplete. No production service uses this wrapper.
