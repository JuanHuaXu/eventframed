# Incremental sort-key publication v1: frozen single-owner probe

Use one private, long-lived LibraVDB Store and one SQLite sidecar. Never
simultaneously open the same LibraVDB path from another Store: the frozen
[dual-open falsifier](mmm-sort-writer-ownership-v1-results.md) rules that out.
Start with three sortable EventFrames and a full-scan genesis certificate.
Build an ordered SQLite journal of row hashes and a chain root. On process
open, validate every journal row against the durable EventFrame, its exact
sort key and vector; require journal row count equal collection row count,
chain root equal READY and READY LSN equal the latest LibraVDB LSN. This
startup validation may be O(N); the append and capture paths may not be.

For an authorized new event, the gate must hold its local reader/writer lock,
check the current READY LSN, commit through the research-only sortable
receipt writer, validate only the new durable rows, then atomically append
journal rows and update the READY LSN/count/root in SQLite. Release readers
only after the sidecar commit and in-memory certificate match. An exact
duplicate must not commit or advance the marker. Capture remains two LSN
reads plus one marker read and is admitted only at the verified exact LSN.

Finite controls:

1. One and then two authorized new rows remain searchable at their as-of
   times, while a future subsecond row remains excluded. Reopen and verify
   the complete journal and marker.
2. Inject an interruption after LibraVDB receipt and before the SQLite
   transaction. Current and reopened gates must deny; no new READY claim.
3. Inject an interruption after the SQLite transaction and before in-memory
   publication. The current gate must deny; reopen must validate and accept.
4. Tamper one journal hash in the private sidecar. Reopen must deny.
5. A legacy unkeyed write through the same Store must stale READY; an
   attempted authorized append must refuse to launder that state.

Measure 100 quiet captures and 100 authorized one-event appends in a fresh
private store after 16 warm-ups; report p50/p99, with a frozen isolated
append p99 ceiling of 100 ms and capture p99 ceiling of 1 ms. Run the
functional controls normally and under `-race`. Passing is a single-owner
component only: no multiprocess fencing, full power-loss guarantee,
automatic recovery from an interrupted DB-only commit, loaded Recall p99,
or Goal 6 completion is inherited.
