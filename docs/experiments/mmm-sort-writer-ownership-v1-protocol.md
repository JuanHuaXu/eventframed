# Sort-key writer ownership v1: frozen private bypass probe

Create a private three-row sortable EventFrame collection and publish the
test-only READY marker. Acquire a nonblocking exclusive advisory file lock
on a separate owner file. A second file descriptor must be unable to acquire
that lock while the first owns it; after release, it must succeed.

While the first owner and Store remain open, try a second independent Store
on the same private LibraVDB path. Record whether Open succeeds. If it does,
write one legacy unkeyed event through that Store and close it before any
further first-Store operation. Record the first Store's latest LSN and READY
gate result. A second Store write that succeeds while the owner lock is held
is an advisory-lock bypass: the lock is not an enforced database fence.
After any successful bypass, an accepted current LSN that includes the
unkeyed row is a fail-open safety error. Acceptance at the old LSN is an
as-of-safe but stale live read and fails Goal 6 freshness; report those
cases separately. Close/reopen the first Store and require
that the marker deny before any full rescan or new authorization. Record
whether the direct write survived reopen. No production path or data is
used. Run the focused test normally and under `-race`.

A failure to observe the external write in an already-open Store is a
negative safety finding, not a reason to weaken the READY check. Even a
successful fail-closed response does not make advisory locking sufficient:
only cooperating writers obey it. This probe does not implement an
incrementally renewed marker, multi-process safety or Goal 6 completion.
