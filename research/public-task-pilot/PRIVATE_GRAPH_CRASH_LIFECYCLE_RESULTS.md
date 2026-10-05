# Private graph crash and shutdown checks

## Abrupt insertion recovery

`TestPrivateGraphAbruptRecovery` starts an isolated child with a fresh temporary
libravdb database and one durable survivor. The new graph coordinator prepares
a second node and enters the real transaction callback. The parent waits for an
explicit marker, then kills only that child process without database Close.

Two barriers: after staging the inserted vector but before revision/commit;
and after WithTx succeeds but before the callback returns or graph publishes.
Each repeated three times under race detection (1.640s for all six kills).

- Staged: reopen yields only the original vector and revision 1.
- Committed: reopen yields both original and inserted vectors and revision 2.
- Survivor vector equality and inserted presence/content are checked explicitly.
- Timeout-triggered exits do not count as controlled crashes; child is reaped.

This is process-failure evidence at two explicit boundaries, not machine power
loss, filesystem corruption, random crash timing or full ANN reconstruction.

## Shutdown

Added context-aware `PrivateGraphWriter.Close`. It takes the serialized writer
gate, so cannot close storage ownership during a callback. Active leases make
Close return capacity without changing admission. After all leases release,
successful Close clears this owner's roots/maps/callback and rejects new leases
and writes. Repeated close is harmless. External graph aliases and the database
remain caller-owned; callers must close the database separately afterward.

Tests cover a live lease across publication, retry after release, idempotence,
post-close insert/delete/acquire, canceled close and close waiting on a blocked
persistence callback. Combined private graph transaction, lease, crash and
lifecycle tests passed three race repetitions (2.177s). Ordinary researchindex
suite passed (5.185s).

## Remaining

Retained-byte bounds and idle-lease expiry are still absent. Existing leases
cannot be forcibly invalidated without changing the API's ownership guarantee.
These lifecycle tests are not sustained concurrent serving measurements. Next
run the original offered-load schedule through the private graph coordinator,
recording queue/preparation/persistence/search times and durable acknowledgements.
All seven whole research goals remain open.
